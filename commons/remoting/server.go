package remoting

import (
	"github.com/sirupsen/logrus"
	"net"
	"strings"
	"sync"
	"time"
)

type RemotingServer struct {
	config  *RemotingConfig
	clients map[string]RemotingChannel

	exitChanOne *sync.Once
	exitChan    chan struct{} // notify all goroutines to shutdown

	waitGroup      *sync.WaitGroup // wait for all goroutines
	coderFactory   RemotingCoderFactory
	handlerFactory RemotingHandlerFactory
}

func (this *RemotingServer) SetCoderFactory(coderFactory RemotingCoderFactory) *RemotingServer {
	this.coderFactory = coderFactory
	return this
}

func (this *RemotingServer) SetCoder(coder RemotingCoder) *RemotingServer {
	return this.SetCoderFactory(func(channel RemotingChannel, config RemotingConfig) RemotingCoder {
		return coder
	})
}

func (this *RemotingServer) SetHandlerFactory(handlerFactory RemotingHandlerFactory) *RemotingServer {
	this.handlerFactory = handlerFactory
	return this
}

func (this *RemotingServer) SetHandler(handler RemotingHandler) *RemotingServer {
	return this.SetHandlerFactory(func(channel RemotingChannel, config RemotingConfig) RemotingHandler {
		return handler
	})
}

func (this *RemotingServer) Start() error {
	if this.coderFactory == nil {
		return ErrNoCoder
	}
	if this.handlerFactory == nil {
		return ErrNoHandler
	}

	if tcpAddr, err := net.ResolveTCPAddr("tcp4", this.config.Listen); err != nil {
		return err
	} else if listener, err := net.ListenTCP("tcp", tcpAddr); err != nil {
		return err
	} else {
		go this.startListener(listener)
		return nil
	}
}
func (this *RemotingServer) startListener(listener *net.TCPListener) {
	this.waitGroup.Add(1)
	defer func() {
		_ = listener.Close()
		this.waitGroup.Done()
		this.Shutdown()
	}()
	logrus.Infof("service startup：%s", listener.Addr().String())

	acceptTimeout := time.Second * time.Duration(this.config.AcceptTimeout)
	for {
		_ = listener.SetDeadline(time.Now().Add(acceptTimeout))
		select {
		case <-this.exitChan:
			return
		default:
			conn, err := listener.AcceptTCP()
			if err != nil {
				if strings.Contains(err.Error(), "i/o timeout") {
					continue
				} else {
					logrus.Errorf("Service monitoring error：%s", err)
					return
				}
			}
			channel := this.newChannel(conn)
			addr := channel.RemoteAddr()
			this.clients[addr] = channel
		}
	}
}

func (this *RemotingServer) newChannel(conn *net.TCPConn) RemotingChannel {
	this.waitGroup.Add(1)

	addr := conn.RemoteAddr().String()
	logrus.Infof("Client connection server：%s", addr)

	channel := NewChannel(conn, this.config)
	channel.waitGroup = this.waitGroup
	channel.coder = this.coderFactory(channel, *this.config)
	channel.handler = this.handlerFactory(channel, *this.config)

	channel.Do(func(ch RemotingChannel) {
		logrus.Debugf("Client close：%s", ch)
		delete(this.clients, ch.RemoteAddr())
		this.waitGroup.Done()
	})
	return channel
}

func (this *RemotingServer) closeChannels() {
	for _, v := range this.clients {
		if v != nil {
			v.Close()
		}
	}
}

func (this *RemotingServer) Shutdown() {
	this.exitChanOne.Do(func() {
		logrus.Infof("Turn off the server")
		close(this.exitChan)
		this.closeChannels()
		logrus.Infof("Service has stopped.")
	})
	this.waitGroup.Wait()
}

func NewRemotingServer(config *RemotingConfig) (*RemotingServer, error) {
	if config == nil {
		config = DefaultConfig()
	}
	server := &RemotingServer{
		config:   config,
		clients:  make(map[string]RemotingChannel),
		exitChan: make(chan struct{}), exitChanOne: &sync.Once{},
		waitGroup: &sync.WaitGroup{},
	}
	return server, nil
}
