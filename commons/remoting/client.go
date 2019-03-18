package remoting

import (
	"github.com/ihaiker/tenured-go-server/commons"
	"net"
	"sync"
	"time"
)

type RemotingClient struct {
	lock sync.Locker
	remotingImpl
}

func (this *RemotingClient) Start() error {
	if err := this.remotingImpl.Start(); err != nil {
		return nil
	}
	this.channelSelector = this.getChannel
	return nil
}

func (this *RemotingClient) getChannel(address string, timeout time.Duration) (RemotingChannel, error) {
	if channel, err := this.remotingImpl.getChannel(address, timeout); err == nil {
		return channel, nil
	} else if noChannel, ok := err.(*RemotingError); ok && noChannel.Op == ErrNoChannel {
		return this.createNewChannel(address, timeout)
	} else {
		return nil, noChannel
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
		channel, err := this.newChannel(address, conn.(*net.TCPConn))
		return channel, err
	}
}

func NewRemotingClient(config *RemotingConfig) *RemotingClient {
	if config == nil {
		config = DefaultConfig()
	}
	client := &RemotingClient{
		lock: &sync.Mutex{},
		remotingImpl: remotingImpl{
			config:    config,
			channels:  make(map[string]RemotingChannel),
			exitChan:  make(chan struct{}),
			status:    commons.S_STATUS_INIT,
			waitGroup: &sync.WaitGroup{},
			hocks:     map[Hock]func(){},
		},
	}
	return client
}
