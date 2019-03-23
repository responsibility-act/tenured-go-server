package protocol

import (
	"errors"
	"github.com/ihaiker/tenured-go-server/commons"
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

type TenuredService interface {
	commons.Service

	Invoke(channel string, command *TenuredCommand, timeout time.Duration) (*TenuredCommand, error)

	AsyncInvoke(channel string, command *TenuredCommand, timeout time.Duration,
		callback func(tenuredCommand *TenuredCommand, err error))

	RegisterCommandProcesser(code uint16, processer TenuredCommandProcesser, executorService executors.ExecutorService)

	IsActive() bool
}

type tenuredService struct {
	remoting         remoting.Remoting
	responseTables   map[uint32]*responseTableBlock
	commandProcesser map[uint16]*tenuredCommandRunner
	*remoting.HandlerWrapper
}

func (this *tenuredService) Invoke(channel string, command *TenuredCommand, timeout time.Duration) (*TenuredCommand, error) {
	if !this.remoting.IsActive() {
		return nil, &TenuredError{Code: remoting.ErrClosed.String(), Message: "closed"}
	}
	requestId := command.id
	responseFuture := future.Set()
	this.responseTables[requestId] = &responseTableBlock{address: channel, future: responseFuture}

	if err := this.remoting.SendTo(channel, command, timeout); err != nil {
		logrus.Debugf("send %d error: %v", requestId, err)
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

func (this *tenuredService) AsyncInvoke(channel string, command *TenuredCommand, timeout time.Duration,
	callback func(tenuredCommand *TenuredCommand, err error)) {

	if !this.remoting.IsActive() {
		callback(nil, &TenuredError{Code: remoting.ErrClosed.String(), Message: "closed"})
		return
	}
	requestId := command.id
	responseFuture := future.Set()
	this.responseTables[requestId] = &responseTableBlock{address: channel, future: responseFuture}

	this.remoting.SyncSendTo(channel, command, timeout, func(err error) {
		if err != nil {
			logrus.Debugf("async send %d error", requestId)
			callback(nil, err)
			delete(this.responseTables, requestId)
		} else {
			logrus.Debugf("async send %d error", requestId)
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

func (this *tenuredService) RegisterCommandProcesser(code uint16, processer TenuredCommandProcesser, executorService executors.ExecutorService) {
	this.commandProcesser[code] = &tenuredCommandRunner{process: processer, executorService: executorService}
}

func (this *tenuredService) makeAck(channel remoting.RemotingChannel, requestCommand *TenuredCommand, header interface{}, err *TenuredError) {
	response := NewACK(requestCommand.id)
	if err != nil {
		response.RemotingError(err)
	}
	if header != nil {
		_ = response.SetHeader(header)
	}
	if err := channel.Write(response, time.Second*7); err != nil {
		if remoting.IsRemotingError(err, remoting.ErrClosed) {
			return
		}
		logrus.Warnf("send ack error: %s", err.Error())
	}
}

func (this *tenuredService) onCommandProcesser(channel remoting.RemotingChannel, command *TenuredCommand) {
	if command.code == REQUEST_CODE_IDLE {
		logrus.Debug("receiver idle ", channel.RemoteAddr())
		this.makeAck(channel, command, nil, nil)
		return
	} else if processRunner, has := this.commandProcesser[command.code]; has {
		processRunner.onCommand(channel, command)
	} else {
		logrus.Warn("not found coder processer:", command.code)
	}
}

func (this *tenuredService) OnMessage(channel remoting.RemotingChannel, msg interface{}) {
	command := msg.(*TenuredCommand)
	if command.IsACK() {
		requestId := command.id
		if f, has := this.responseTables[requestId]; has {
			f.future.Set(command)
		}
		return
	} else {
		this.onCommandProcesser(channel, command)
	}
}

//发送心跳包
func (this *tenuredService) OnIdle(channel remoting.RemotingChannel) {
	if err := channel.Write(NewIdle(), time.Second*3); err != nil {
		if remoting.IsRemotingError(err, remoting.ErrClosed) {
			return
		}
		logrus.Warnf("send %s idle error: %v", channel.RemoteAddr(), err)
	}
}

func (this *tenuredService) OnClose(channel remoting.RemotingChannel) {
	this.fastFailChannel(channel)
}

func (this *tenuredService) fastFailChannel(channel remoting.RemotingChannel) {
	for _, v := range this.responseTables {
		if v.address == channel.RemoteAddr() {
			v.future.Exception(errors.New(remoting.ErrClosed.String()))
		}
	}
}

func (this *tenuredService) Start() error {
	return this.remoting.Start()
}

func (this *tenuredService) waitRequest(interrupt bool) {
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

func (this *tenuredService) IsActive() bool {
	return this.remoting.IsActive()
}

func (this *tenuredService) IsStatus(status commons.ServerStatus) bool {
	return this.remoting.IsStatus(status)
}

func (this *tenuredService) Shutdown(interrupt bool) {
	this.remoting.RegisterHock(remoting.HOCK_SHUTDOWN_AFTER, func() {
		this.waitRequest(interrupt)
	})
	this.remoting.Shutdown(interrupt)
}
