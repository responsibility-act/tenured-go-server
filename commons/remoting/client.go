package remoting

import (
	"github.com/ihaiker/tenured-go-server/commons"
	"net"
	"sync"
	"time"
)

type RemotingClient struct {
	lock sync.Locker
	Remoting
}

func (this *RemotingClient) Start() error {
	if err := this.Remoting.Start(); err != nil {
		return nil
	}
	this.channelSelector = this.getChannel
	return nil
}

func (this *RemotingClient) getChannel(address string, timeout time.Duration) (RemotingChannel, error) {
	if channel, err := this.Remoting.getChannel(address, timeout); err == nil {
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
		channel := this.newChannel(address, conn.(*net.TCPConn))
		return channel, nil
	}
}

func NewClient(config *RemotingConfig) *RemotingClient {
	if config == nil {
		config = DefaultConfig()
	}
	client := &RemotingClient{
		lock: &sync.Mutex{},
		Remoting: Remoting{
			config:    config,
			channels:  make(map[string]RemotingChannel),
			exitChan:  make(chan struct{}),
			status:    commons.S_STATUS_INIT,
			waitGroup: &sync.WaitGroup{},
		},
	}
	return client
}
