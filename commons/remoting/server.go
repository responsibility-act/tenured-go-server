package remoting

import (
	"github.com/ihaiker/tenured-go-server/commons"
	"github.com/sirupsen/logrus"
	"net"
	"sync"
	"time"
)

type RemotingServer struct {
	address string
	remotingImpl
}

func (this *RemotingServer) Start() error {
	if err := this.remotingImpl.Start(); err != nil {
		return nil
	}

	if tcpAddr, err := net.ResolveTCPAddr("tcp4", this.address); err != nil {
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
		this.Shutdown(false)
	}()
	logrus.Infof("server startup：%s", listener.Addr().String())

	acceptTimeout := time.Second * time.Duration(this.config.AcceptTimeout)
	for {
		_ = listener.SetDeadline(time.Now().Add(acceptTimeout))
		select {
		case <-this.exitChan:
			return
		default:
			conn, err := listener.AcceptTCP()
			if err != nil {
				if netErr, ok := err.(*net.OpError); ok && netErr.Timeout() {
					continue
				} else {
					logrus.Errorf("Service monitoring error：%s", err)
					return
				}
			}
			address := conn.RemoteAddr().String()
			if _, err = this.newChannel(address, conn); err != nil {
				logrus.Infof("the server reject connection. %s", err.Error())
			}
		}
	}
}
func NewRemotingServer(address string, config *RemotingConfig) (*RemotingServer, error) {
	if config == nil {
		config = DefaultConfig()
	}
	server := &RemotingServer{
		address: address,
		remotingImpl: remotingImpl{
			config:    config,
			channels:  make(map[string]RemotingChannel),
			exitChan:  make(chan struct{}),
			status:    commons.S_STATUS_INIT,
			waitGroup: &sync.WaitGroup{},
			hocks:     map[Hock]func(){},
		},
	}
	return server, nil
}
