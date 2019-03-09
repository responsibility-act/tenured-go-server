package remoting

import (
	"github.com/sirupsen/logrus"
	"net"
	"strings"
	"sync"
	"time"
)

type remotingServer struct {
	config  *remotingConfig
	clients map[string]Channel

	exitChanOne *sync.Once
	exitChan    chan struct{} // notify all goroutines to shutdown

	waitGroup      *sync.WaitGroup // wait for all goroutines
	coderFactory   CoderFactory
	handlerFactory HandlerFactory
}

func (this *remotingServer) SetCoderFactory(coderFactory func(Channel) Coder) *remotingServer {
	this.coderFactory = coderFactory
	return this
}

func (this *remotingServer) SetHandlerFactory(handlerFactory func(Channel) Handler) *remotingServer {
	this.handlerFactory = handlerFactory
	return this
}

func (this *remotingServer) Start() error {
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
func (this *remotingServer) startListener(listener *net.TCPListener) {
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

func (this *remotingServer) newChannel(conn *net.TCPConn) Channel {
	this.waitGroup.Add(1)

	addr := conn.RemoteAddr().String()
	logrus.Infof("Client connection server：%s", addr)

	channel := NewChannel(conn, this.config)
	channel.waitGroup = this.waitGroup
	channel.coder = this.coderFactory(channel)
	channel.handler = this.handlerFactory(channel)

	channel.Do(func(ch Channel) {
		logrus.Debugf("Client close：%s", ch)
		delete(this.clients, ch.RemoteAddr())
		this.waitGroup.Done()
	})
	return channel
}

func (this *remotingServer) closeChannels() {
	for _, v := range this.clients {
		if v != nil {
			v.Close()
		}
	}
}

func (this *remotingServer) Shutdown() {
	this.exitChanOne.Do(func() {
		logrus.Infof("Turn off the server")
		close(this.exitChan)
		this.closeChannels()
		logrus.Infof("Service has stopped.")
	})
	this.waitGroup.Wait()
}

func NewRemotingServer(config *remotingConfig) (error, *remotingServer) {
	if config == nil {
		config = DefaultConfig()
	}
	server := &remotingServer{
		config:   config,
		clients:  make(map[string]Channel),
		exitChan: make(chan struct{}), exitChanOne: &sync.Once{},
		waitGroup: &sync.WaitGroup{},
	}
	return nil, server
}
