package protocol

import (
	"errors"
	"github.com/ihaiker/tenured-go-server/commons/executors"
	"github.com/ihaiker/tenured-go-server/commons/future"
	"github.com/ihaiker/tenured-go-server/commons/remoting"
	"github.com/sirupsen/logrus"
	"reflect"
	"time"
)

type responseTableBlock struct {
	address string
	future  *future.SetFuture
}

type TenuredServer struct {
	server           *remoting.RemotingServer
	responseTables   map[uint32]*responseTableBlock
	commandProcesser map[uint16]*tenuredCommandRunner
	*remoting.HandlerWrapper

	closeStatus uint32
}

func (this *TenuredServer) Invoke(channel string, command *TenuredCommand, timeout time.Duration) (*TenuredCommand, error) {
	if !this.server.GetStatus().IsUp() {
		return nil, &TenuredError{Code: remoting.ErrClosed.String(), Message: "closed"}
	}
	requestId := command.Id
	responseFuture := future.Set()
	this.responseTables[requestId] = &responseTableBlock{address: channel, future: responseFuture}

	if err := this.server.SendTo(channel, command, timeout); err != nil {
		logrus.Debug("send %d error:", requestId, err)
		delete(this.responseTables, requestId)
		return nil, err
	} else {
		response, err := responseFuture.GetWithTimeout(timeout)
		delete(this.responseTables, requestId)
		if err != nil {
			return nil, err
		}
		if responseCommand, match := response.(*TenuredCommand); !match {
			return nil, errors.New("response type error：" + reflect.TypeOf(response).Name())
		} else {
			return responseCommand, nil
		}
	}
}

func (this *TenuredServer) AsyncInvoke(channel string, command *TenuredCommand, timeout time.Duration,
	callback func(tenuredCommand *TenuredCommand, err error)) {
	if !this.server.GetStatus().IsUp() {
		callback(nil, &TenuredError{Code: remoting.ErrClosed.String(), Message: "closed"})
		return
	}
	requestId := command.Id
	responseFuture := future.Set()
	this.responseTables[requestId] = &responseTableBlock{address: channel, future: responseFuture}

	this.server.SyncSendTo(channel, command, timeout, func(err error) {
		if err != nil {
			logrus.Debug("async send %d error", requestId)
			callback(nil, err)
			delete(this.responseTables, requestId)
		} else {
			logrus.Debug("async send %d error", requestId)
		}
	})

	go func() {
		response, err := responseFuture.GetWithTimeout(timeout)
		delete(this.responseTables, requestId)

		if err != nil {
			callback(nil, err)
			return
		}

		if responseCommand, match := response.(*TenuredCommand); !match {
			callback(nil, errors.New("response type error："+reflect.TypeOf(response).Name()))
		} else {
			callback(responseCommand, nil)
		}
	}()
}

func (this *TenuredServer) RegisterCommandProcesser(code uint16, processer TenuredCommandProcesser, executorService executors.ExecutorService) {
	this.commandProcesser[code] = &tenuredCommandRunner{process: processer, executorService: executorService}
}

func (this *TenuredServer) makeAck(channel remoting.RemotingChannel, command *TenuredCommand) {
	response := NewACK(command.Id)
	if err := channel.Write(response, time.Second*7); err != nil {
		if remoting.IsRemotingError(err, remoting.ErrClosed) {
			return
		}
		logrus.Warn("send ack error: %s", err.Error())
	}
}

func (this *TenuredServer) onCommandProcesser(channel remoting.RemotingChannel, command *TenuredCommand) {
	this.makeAck(channel, command)
	if processRunner, has := this.commandProcesser[command.Code]; has {
		processRunner.onCommand(channel, command)
	}
}

func (this *TenuredServer) OnMessage(channel remoting.RemotingChannel, msg interface{}) {
	command := msg.(*TenuredCommand)
	if command.IsACK() {
		requestId := command.Id
		if f, has := this.responseTables[requestId]; has {
			f.future.Set(command)
		}
		return
	} else {
		this.onCommandProcesser(channel, command)
	}
}

//发送心跳包
func (this *TenuredServer) OnIdle(channel remoting.RemotingChannel) {
	if err := channel.Write(NewIdle(), time.Second*3); err != nil {
		if remoting.IsRemotingError(err, remoting.ErrClosed) {
			return
		}
		logrus.Warnf("send %s idle error: %v", channel.RemoteAddr(), err)
	}
}

func (this *TenuredServer) OnClose(channel remoting.RemotingChannel) {
	this.fastFailChannel(channel)
}

func (this *TenuredServer) fastFailChannel(channel remoting.RemotingChannel) {
	for _, v := range this.responseTables {
		if v.address == channel.RemoteAddr() {
			v.future.Exception(errors.New(remoting.ErrClosed.String()))
		}
	}
}

func (this *TenuredServer) Start() error {
	return this.server.Start()
}

func (this *TenuredServer) waitRequest(interrupt bool) {
	if interrupt {
		for _, v := range this.responseTables {
			v.future.Exception(errors.New(remoting.ErrClosed.String()))
		}
	} else {
		for {
			if len(this.responseTables) == 0 {
				return
			}
			<-time.After(time.Millisecond * 200)
		}
	}
}

func (this *TenuredServer) Shutdown(interrupt bool) {
	this.server.Shutdown(interrupt)
}

func NewTenuredServer(address string, config *remoting.RemotingConfig) (*TenuredServer, error) {
	if remotingServer, err := remoting.NewRemotingServer(address, config); err != nil {
		return nil, err
	} else {
		remotingServer.SetCoder(&tenuredCoder{config: config})
		server := &TenuredServer{
			server:           remotingServer,
			responseTables:   map[uint32]*responseTableBlock{},
			commandProcesser: map[uint16]*tenuredCommandRunner{},
		}
		remotingServer.SetHandler(server)
		return server, nil
	}
}
