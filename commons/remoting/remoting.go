package remoting

import (
	"errors"
	"github.com/ihaiker/tenured-go-server/commons"
	"github.com/sirupsen/logrus"
	"net"
	"sync"
	"time"
)

type Hock int

const (
	HOCK_START_BEFORE Hock = iota
	HOCK_START_AFTER

	HOCK_SHUTDOWN_BEFORE
	HOCK_SHUTDOWN_AFTER
)

type Remoting interface {
	commons.Service

	SetCoderFactory(coderFactory RemotingCoderFactory)
	SetCoder(coder RemotingCoder)
	SetHandlerFactory(handlerFactory RemotingHandlerFactory)
	SetHandler(handler RemotingHandler)
	RegisterHock(hock Hock, fn func())

	SendTo(address string, msg interface{}, timeout time.Duration) error
	SyncSendTo(address string, msg interface{}, timeout time.Duration, callback func(error))

	IsActive() bool
	IsStatus(status commons.ServerStatus) bool
}

type remotingImpl struct {
	config   *RemotingConfig
	channels map[string]RemotingChannel

	status   commons.ServerStatus
	exitChan chan struct{} // notify all goroutines to shutdown

	waitGroup      *sync.WaitGroup // wait for all goroutines
	coderFactory   RemotingCoderFactory
	handlerFactory RemotingHandlerFactory

	hocks           map[Hock]func()
	channelSelector func(address string, timeout time.Duration) (RemotingChannel, error)
}

func (this *remotingImpl) SetCoderFactory(coderFactory RemotingCoderFactory) {
	this.coderFactory = coderFactory
}

func (this *remotingImpl) SetCoder(coder RemotingCoder) {
	this.SetCoderFactory(func(channel RemotingChannel, config RemotingConfig) RemotingCoder {
		return coder
	})
}

func (this *remotingImpl) SetHandlerFactory(handlerFactory RemotingHandlerFactory) {
	this.handlerFactory = handlerFactory
}

func (this *remotingImpl) SetHandler(handler RemotingHandler) {
	this.SetHandlerFactory(func(channel RemotingChannel, config RemotingConfig) RemotingHandler {
		return handler
	})
}

func (this *remotingImpl) RegisterHock(hock Hock, fn func()) {
	this.hocks[hock] = fn
}

func (this *remotingImpl) Start() error {
	if this.coderFactory == nil {
		return &RemotingError{Op: ErrCoder, Err: errors.New("no coder")}
	}
	if this.handlerFactory == nil {
		return &RemotingError{Op: ErrHandler, Err: errors.New("no handler")}
	}
	this.channelSelector = this.getChannel

	if hock, has := this.hocks[HOCK_START_BEFORE]; has {
		hock()
	}
	if started := this.status.Start(func() {}); started {
		if hock, has := this.hocks[HOCK_START_AFTER]; has {
			hock()
		}
		return nil
	} else {
		return errors.New("start unknow error!")
	}
}

func (this *remotingImpl) SendTo(address string, msg interface{}, timeout time.Duration) error {
	timeoutTime := time.Now().Add(timeout)
	if channel, err := this.channelSelector(address, timeout); err != nil {
		return err
	} else {
		if timeoutTime.Before(time.Now()) {
			return &RemotingError{Op: ErrSendTimeout, Err: errors.New("send timeout")}
		}
		return channel.Write(msg, timeout)
	}
}

func (this *remotingImpl) SyncSendTo(address string, msg interface{}, timeout time.Duration, callback func(error)) {
	timeoutTime := time.Now().Add(timeout)
	if channel, err := this.channelSelector(address, timeout); err != nil {
		callback(err)
	} else {
		if timeoutTime.Before(time.Now()) {
			callback(&RemotingError{Op: ErrSendTimeout, Err: errors.New("send timeout")})
			return
		}
		channel.AsyncWrite(msg, timeout, callback)
	}
}

func (this *remotingImpl) getChannel(address string, timeout time.Duration) (RemotingChannel, error) {
	if channel, ok := this.channels[address]; ok {
		return channel, nil
	} else {
		return nil, &RemotingError{Op: ErrNoChannel, Err: errors.New("not found channel " + address)}
	}
}

func (this *remotingImpl) newChannel(address string, conn *net.TCPConn) (RemotingChannel, error) {
	this.waitGroup.Add(1)
	logrus.Debugf("new channel：%s", address)

	channel := NewChannel(conn, this.config)
	channel.addr = address
	channel.waitGroup = this.waitGroup
	channel.coder = this.coderFactory(channel, *this.config)
	channel.handler = this.handlerFactory(channel, *this.config)
	this.channels[address] = channel
	err := channel.Do(func(ch RemotingChannel) {
		delete(this.channels, ch.RemoteAddr())
		this.waitGroup.Done()
	})
	return channel, err
}

func (this *remotingImpl) IsActive() bool {
	return this.status.IsUp()
}

func (this *remotingImpl) IsStatus(status commons.ServerStatus) bool {
	return this.status.Is(status)
}

func (this *remotingImpl) closeChannels() {
	for _, v := range this.channels {
		if v != nil {
			v.Close()
		}
	}
}

func (this *remotingImpl) Shutdown(interrupt bool) {
	this.status.Shutdown(func() {
		this.waitGroup.Add(1)
		logrus.Infof("turn off remoting")
		if hock, has := this.hocks[HOCK_SHUTDOWN_BEFORE]; has {
			hock()
		}
		close(this.exitChan)
		if hock, has := this.hocks[HOCK_SHUTDOWN_AFTER]; has {
			hock()
		}
		this.closeChannels()
		logrus.Infof("remoting has stopped.")
		this.waitGroup.Done()
	})
	this.waitGroup.Wait()
}
