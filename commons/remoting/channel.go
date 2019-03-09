package remoting

import (
	"github.com/ihaiker/tenured-go-server/commons"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"strings"
	"sync"
	"time"
)

type Channel interface {
	RemoteAddr() string

	ChannelAttributes() map[string]string

	Write(msg interface{}, timeout time.Duration) error

	Close()
}

type defChannel struct {
	config *remotingConfig

	addr    string
	conn    *net.TCPConn
	coder   Coder
	handler Handler

	attributes map[string]string

	onCloseFn func(channel Channel)

	closeChan chan struct{}
	closeOnce *sync.Once

	waitGroup   *sync.WaitGroup
	idleTimer   *time.Timer
	idleTimeout int
}

func (this *defChannel) RemoteAddr() string {
	return this.addr
}
func (this *defChannel) ChannelAttributes() map[string]string {
	return this.attributes
}
func (this *defChannel) Write(msg interface{}, timeout time.Duration) error {
	var err error = nil
	commons.Try(func() {
		if err := this.coder.Encode(msg, this.conn); err != nil {
			panic(err)
		}
	}, func(e error) {
		err = e
	})
	return err
}
func (this *defChannel) Do(onClose func(channel Channel)) {
	this.onCloseFn = onClose
	go this.syncDo(this.readLoop)
	go this.syncDo(this.heartbeatLoop)
	this.handler.OnChannel(this)
}

func (this *defChannel) Close() {
	this.closeOnce.Do(func() {
		this.handler.OnClose(this)

		if this.onCloseFn != nil {
			this.onCloseFn(this)
		}
		close(this.closeChan)
	})
}

func (this *defChannel) readLoop() {
	defer func() {
		_ = this.conn.Close()
		this.Close()
	}()

	for {
		select {
		case <-this.closeChan:
			return
		default:
			if msg, err := this.coder.Decode(this.conn); err != nil {
				if !this.isClosed(err) {
					this.handler.OnError(this, ErrDecoder, err)
				}
				return
			} else {
				this.resetReadIdle()
				commons.Try(func() {
					this.handler.OnMessage(this, msg)
				}, func(err error) {
					this.handler.OnError(this, err, msg)
				})
			}
		}
	}
}

func (c *defChannel) isClosed(err error) bool {
	if strings.Contains(err.Error(), "connection reset by peer") {
		return true
	} else if err == io.EOF {
		return true
	}
	return false
}

func (this *defChannel) resetReadIdle() {
	this.idleTimer.Reset(time.Millisecond * time.Duration(this.config.IdleTime))
	this.idleTimeout = 0
}

func (this *defChannel) heartbeatLoop() {
	logrus.Debug("start heartbeat loop:", this.RemoteAddr())
	defer func() {
		this.idleTimer.Stop()
		logrus.Info("close heartbeat:", this.RemoteAddr())
		this.Close()
	}()

	idleCheckTime := time.Millisecond * time.Duration(this.config.IdleTime)
	for {
		select {
		case <-this.closeChan:
			return
		case <-this.idleTimer.C:
			if this.idleTimeout+1 <= this.config.IdleTimeout {
				this.idleTimeout = this.idleTimeout + 1
				logrus.Info("send idle ", this.RemoteAddr(), ", time:", this.idleTimeout)
				this.idleTimer.Reset(idleCheckTime)
				this.handler.OnIdle(this)
			} else {
				logrus.Info("IdleTimerOut: ", this.RemoteAddr())
				return
			}
		}
	}
}

func (this *defChannel) syncDo(fn func()) {
	this.waitGroup.Add(1)
	defer this.waitGroup.Done()
	fn()
}

func NewChannel(conn *net.TCPConn, config *remotingConfig) *defChannel {
	channel := &defChannel{
		config:     config,
		conn:       conn,
		addr:       conn.RemoteAddr().String(),
		attributes: make(map[string]string),
		closeChan:  make(chan struct{}),
		closeOnce:  &sync.Once{},
	}
	channel.idleTimer = time.NewTimer(time.Millisecond * time.Duration(config.IdleTime))
	channel.idleTimeout = 0
	return channel
}
