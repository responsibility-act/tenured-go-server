package remoting

import (
	"errors"
	"github.com/ihaiker/tenured-go-server/commons"
	"github.com/sirupsen/logrus"
	"net"
	"sync"
	"time"
)

type Remoting struct {
	config   *RemotingConfig
	channels map[string]RemotingChannel

	status   commons.ServerStatus
	exitChan chan struct{} // notify all goroutines to shutdown

	waitGroup      *sync.WaitGroup // wait for all goroutines
	coderFactory   RemotingCoderFactory
	handlerFactory RemotingHandlerFactory

	channelSelector func(address string, timeout time.Duration) (RemotingChannel, error)
}

func (this *Remoting) SetCoderFactory(coderFactory RemotingCoderFactory) {
	this.coderFactory = coderFactory
}

func (this *Remoting) SetCoder(coder RemotingCoder) {
	this.SetCoderFactory(func(channel RemotingChannel, config RemotingConfig) RemotingCoder {
		return coder
	})
}

func (this *Remoting) SetHandlerFactory(handlerFactory RemotingHandlerFactory) {
	this.handlerFactory = handlerFactory
}

func (this *Remoting) SetHandler(handler RemotingHandler) {
	this.SetHandlerFactory(func(channel RemotingChannel, config RemotingConfig) RemotingHandler {
		return handler
	})
}

func (this *Remoting) Start() error {
	if this.coderFactory == nil {
		return &RemotingError{Op: ErrCoder, Err: errors.New("no coder")}
	}
	if this.handlerFactory == nil {
		return &RemotingError{Op: ErrHandler, Err: errors.New("no handler")}
	}
	this.channelSelector = this.getChannel

	if started := this.status.Start(func() {}); started {
		return nil
	} else {
		return errors.New("start unknow error!")
	}
}

func (this *Remoting) SendTo(address string, msg interface{}, timeout time.Duration) error {
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

func (this *Remoting) SyncSendTo(address string, msg interface{}, timeout time.Duration, callback func(error)) {
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

func (this *Remoting) getChannel(address string, timeout time.Duration) (RemotingChannel, error) {
	if channel, ok := this.channels[address]; ok {
		return channel, nil
	} else {
		return nil, &RemotingError{Op: ErrNoChannel, Err: errors.New("not found channel " + address)}
	}
}

func (this *Remoting) newChannel(address string, conn *net.TCPConn) RemotingChannel {
	this.waitGroup.Add(1)
	logrus.Infof("new channel：%s", address)

	channel := NewChannel(conn, this.config)
	channel.addr = address
	channel.waitGroup = this.waitGroup
	channel.coder = this.coderFactory(channel, *this.config)
	channel.handler = this.handlerFactory(channel, *this.config)
	this.channels[address] = channel

	channel.Do(func(ch RemotingChannel) {
		delete(this.channels, ch.RemoteAddr())
		this.waitGroup.Done()
	})
	return channel
}

func (this *Remoting) GetStatus() commons.ServerStatus {
	return this.status
}

func (this *Remoting) closeChannels() {
	for _, v := range this.channels {
		if v != nil {
			v.Close()
		}
	}
}

func (this *Remoting) Shutdown(interrupt bool) {
	this.status.Shutdown(func() {
		this.waitGroup.Add(1)
		logrus.Infof("turn off remoting")
		close(this.exitChan)
		this.closeChannels()
		logrus.Infof("remoting has stopped.")
		this.waitGroup.Done()
	})
	this.waitGroup.Wait()
}
