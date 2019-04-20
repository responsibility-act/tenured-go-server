package protocol

import (
	"errors"
	"github.com/ihaiker/tenured-go-server/commons"
	"github.com/ihaiker/tenured-go-server/commons/c8tmap"
	"github.com/ihaiker/tenured-go-server/commons/executors"
	"github.com/ihaiker/tenured-go-server/commons/future"
	"github.com/ihaiker/tenured-go-server/commons/remoting"
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
	responseTables   c8tmap.ConcurrentMap //map[uint32]*responseTableBlock，tome: golang map不能并发写入。
	commandProcesser map[uint16]*tenuredCommandRunner
	*remoting.HandlerWrapper
}

func (this *tenuredService) Invoke(channel string, command *TenuredCommand, timeout time.Duration) (*TenuredCommand, error) {
	if !this.remoting.IsActive() {
		return nil, &TenuredError{code: remoting.ErrClosed.String(), message: "closed"}
	}
	requestId := command.id
	responseFuture := future.Set()

	//this.responseTables[requestId] = &responseTableBlock{address: channel, future: responseFuture}
	this.responseTables.Set(requestId, &responseTableBlock{address: channel, future: responseFuture})

	if err := this.remoting.SendTo(channel, command, timeout); err != nil {
		logger.Debugf("send %d error: %v", requestId, err)
		//delete(this.responseTables, requestId)
		this.responseTables.Remove(requestId)
		return nil, err
	} else {
		response, err := responseFuture.GetWithTimeout(timeout)
		//delete(this.responseTables, requestId)
		this.responseTables.Remove(requestId)
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
		callback(nil, &TenuredError{code: remoting.ErrClosed.String(), message: "closed"})
		return
	}
	requestId := command.id
	responseFuture := future.Set()
	//this.responseTables[requestId] = &responseTableBlock{address: channel, future: responseFuture}
	this.responseTables.Set(requestId, &responseTableBlock{address: channel, future: responseFuture})

	this.remoting.SyncSendTo(channel, command, timeout, func(err error) {
		if err != nil {
			logger.Debugf("async send %d error", requestId)
			callback(nil, err)
			this.responseTables.Remove(requestId)
		} else {
			logger.Debugf("async send %d error", requestId)
		}
	})

	//TODO 设置异步执行可调用携程管理
	go func() {
		response, err := responseFuture.GetWithTimeout(timeout)
		//delete(this.responseTables, requestId)
		this.responseTables.Remove(requestId)
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
		logger.Warnf("send ack error: %s", err.Error())
	}
}

func (this *tenuredService) onCommandProcesser(channel remoting.RemotingChannel, command *TenuredCommand) {
	if command.code == REQUEST_CODE_IDLE {
		logger.Debug("receiver idle ", channel.RemoteAddr())
		this.makeAck(channel, command, nil, nil)
		return
	} else if processRunner, has := this.commandProcesser[command.code]; has {
		processRunner.onCommand(channel, command)
	} else {
		logger.Warn("not found process: ", command.code)
	}
}

func (this *tenuredService) OnMessage(channel remoting.RemotingChannel, msg interface{}) {
	command := msg.(*TenuredCommand)
	if command.IsACK() {
		requestId := command.id
		if f, has := this.responseTables.Pop(requestId); has {
			f.(*responseTableBlock).future.Set(command)
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
		logger.Warnf("send %s idle error: %v", channel.RemoteAddr(), err)
	}
}

func (this *tenuredService) OnClose(channel remoting.RemotingChannel) {
	this.fastFailChannel(channel)
}

func (this *tenuredService) fastFailChannel(channel remoting.RemotingChannel) {
	it := this.responseTables.IterBuffered()
	for tu := <-it; tu.Key != nil; tu = <-it {
		if val, has := this.responseTables.Get(tu.Key); has {
			block := val.(*responseTableBlock)
			if block.address == channel.RemoteAddr() {
				block.future.Exception(errors.New(remoting.ErrClosed.String()))
				this.responseTables.Remove(tu.Key)
			}
		}
	}
}

func (this *tenuredService) RegisterHock(hock remoting.Hock, fn func()) {
	this.remoting.RegisterHock(hock, fn)
}

func (this *tenuredService) Start() error {
	return this.remoting.Start()
}

func (this *tenuredService) waitRequest(interrupt bool) {
	if interrupt {
		it := this.responseTables.IterBuffered()
		for tu := <-it; tu.Key != nil; tu = <-it {
			if block, has := this.responseTables.Pop(tu.Key); has {
				block.(*responseTableBlock).future.Exception(errors.New(remoting.ErrClosed.String()))
			}
		}
	} else {
		for {
			if this.responseTables.Count() == 0 {
				return
			}
			<-time.After(time.Millisecond * 10)
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
