package remoting

import (
	"github.com/sirupsen/logrus"
	"net"
	"sync"
	"time"
)

type RemotingClient struct {
	lock         sync.Locker
	loadBalancer RemotingLoadBalancer

	channels map[string]RemotingChannel

	config    *RemotingConfig
	waitGroup *sync.WaitGroup
	closeOnce *sync.Once

	coderFactory   RemotingCoderFactory
	handlerFactory RemotingHandlerFactory
}

func (this *RemotingClient) SetCoderFactory(coderFactory RemotingCoderFactory) *RemotingClient {
	this.coderFactory = coderFactory
	return this
}

func (this *RemotingClient) SetCoder(coder RemotingCoder) *RemotingClient {
	return this.SetCoderFactory(func(channel RemotingChannel, config RemotingConfig) RemotingCoder {
		return coder
	})
}

func (this *RemotingClient) SetHandlerFactory(handlerFactory RemotingHandlerFactory) *RemotingClient {
	this.handlerFactory = handlerFactory
	return this
}

func (this *RemotingClient) SetHandler(handler RemotingHandler) *RemotingClient {
	return this.SetHandlerFactory(func(channel RemotingChannel, config RemotingConfig) RemotingHandler {
		return handler
	})
}

func (this *RemotingClient) newChannel(conn *net.TCPConn) RemotingChannel {
	this.waitGroup.Add(1)
	addr := conn.RemoteAddr().String()
	logrus.Infof("Connection ：%s", addr)

	channel := NewChannel(conn, this.config)

	channel.waitGroup = this.waitGroup
	channel.coder = this.coderFactory(channel, *this.config)
	channel.handler = this.handlerFactory(channel, *this.config)
	this.channels[addr] = channel

	channel.Do(func(ch RemotingChannel) {
		logrus.Debugf("close：%s", ch)
		delete(this.channels, ch.RemoteAddr())
		this.waitGroup.Done()
	})
	return channel
}

func (this *RemotingClient) Send(msg interface{}, timeout time.Duration) error {
	return this.SendTo(this.getDefaultChannel(msg), msg, timeout)
}

func (this *RemotingClient) SyncSend(msg interface{}, timeout time.Duration, callback func(error)) {
	this.SyncSendTo(this.getDefaultChannel(msg), msg, timeout, callback)
}

func (this *RemotingClient) SendTo(address string, msg interface{}, timeout time.Duration) error {
	if channel, err := this.getChannel(address); err != nil {
		return err
	} else {
		return channel.Write(msg, timeout)
	}
}

func (this *RemotingClient) SyncSendTo(address string, msg interface{}, timeout time.Duration, callback func(error)) {
	if channel, err := this.getChannel(address); err != nil {
		callback(err)
	} else {
		channel.AsyncWrite(msg, timeout, callback)
	}
}

func (this *RemotingClient) BeforeClose() {

}

func (this *RemotingClient) AfterClose() {

}

func (this *RemotingClient) closeChannels() {
	for _, channel := range this.channels {
		channel.Close()
	}
}

func (this *RemotingClient) Start() error {
	if this.coderFactory == nil {
		return ErrNoCoder
	}
	if this.handlerFactory == nil {
		return ErrNoHandler
	}
	return nil
}

func (this *RemotingClient) Shutdown() {
	this.closeOnce.Do(func() {
		this.waitGroup.Add(1)
		this.BeforeClose()
		logrus.Info("close ")
		this.closeChannels()
		this.AfterClose()
		this.waitGroup.Done()
	})
	this.waitGroup.Done()
}

func (this *RemotingClient) getDefaultChannel(msg interface{}) string {
	return this.loadBalancer.Selector(msg)
}

func (this *RemotingClient) getChannel(address string) (RemotingChannel, error) {
	if channel, ok := this.channels[address]; ok {
		return channel, nil
	} else {
		return this.createNewChannel(address, time.Second*time.Duration(this.config.ConnectTimeout))
	}
}

func (this *RemotingClient) createNewChannel(address string, timeout time.Duration) (RemotingChannel, error) {
	this.lock.Lock()
	defer this.lock.Unlock()

	if channel, ok := this.channels[address]; ok { //并发的双重检查
		return channel, nil
	}

	if conn, err := net.DialTimeout("tcp", address, timeout); err != nil {
		return nil, err
	} else {
		channel := this.newChannel(conn.(*net.TCPConn))
		return channel, nil
	}
}

func NewClient(config *RemotingConfig, loadBalancer RemotingLoadBalancer) *RemotingClient {
	return &RemotingClient{
		config:       config,
		lock:         &sync.Mutex{},
		loadBalancer: loadBalancer,
		channels:     make(map[string]RemotingChannel),
		closeOnce:    &sync.Once{},
		waitGroup:    &sync.WaitGroup{},
	}
}
