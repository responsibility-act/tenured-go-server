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

type RemotingChannel interface {
	RemoteAddr() string

	ChannelAttributes() map[string]string

	Write(msg interface{}, timeout time.Duration) error

	AsyncWrite(msg interface{}, timeout time.Duration, callback func(error))

	Close()
}

type sendMessage struct {
	msg     []byte
	result  chan error
	timeout time.Time
}

type defChannel struct {
	config *RemotingConfig

	addr    string
	conn    *net.TCPConn
	coder   RemotingCoder
	handler RemotingHandler

	attributes map[string]string

	onCloseFn func(channel RemotingChannel)

	closeChan chan struct{}
	closeOnce *sync.Once

	sendChan chan sendMessage

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

func (this *defChannel) encodeMessage(msg interface{}) (bs []byte, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = commons.Catch(e)
			if !this.isClosed(err) {
				this.handler.OnError(this, ErrEncoder, err)
			}
		}
	}()

	if bs, err = this.coder.Encode(this, msg); err != nil {
		return nil, err
	} else if len(bs) > this.config.PacketBytesLimit {
		return nil, ErrPacketBytesLimit
	} else {
		return bs, nil
	}
}

func (this *defChannel) Write(msg interface{}, timeout time.Duration) error {
	timeoutTime := time.Now().Add(timeout)
	if bs, err := this.encodeMessage(msg); err != nil {
		return err
	} else {
		if timeoutTime.After(time.Now()) { //encode timeout
			return ErrSendTimeout
		}
		result := make(chan error, 1)
		this.sendChan <- sendMessage{msg: bs, timeout: time.Now().Add(timeout), result: result}
		err := <-result
		close(result)
		return err
	}
}
func (this *defChannel) AsyncWrite(msg interface{}, timeout time.Duration, callback func(error)) {
	timeoutTime := time.Now().Add(timeout)
	if bs, err := this.encodeMessage(msg); err != nil {
		callback(err)
		return
	} else {
		if timeoutTime.After(time.Now()) { //encode timeout
			callback(ErrSendTimeout)
			return
		}
		go func() {
			result := make(chan error, 1)
			this.sendChan <- sendMessage{msg: bs, timeout: time.Now().Add(timeout), result: result}
			err := <-result
			close(result)
			callback(err)
		}()
	}
}

func (this *defChannel) Do(onClose func(channel RemotingChannel)) {
	this.onCloseFn = onClose
	go this.syncDo(this.readLoop)
	go this.syncDo(this.heartbeatLoop)
	go this.syncDo(this.writeLoop)
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

func (this *defChannel) closeUnWriteMessageChan() {
	for {
		select {
		case <-time.After(time.Second):
			return
		case msg := <-this.sendChan:
			msg.result <- ErrClosed
		}
	}
}

func (this *defChannel) writeLoop() {
	defer this.closeUnWriteMessageChan()

	for {
		select {
		case <-this.closeChan:
			return
		case msg := <-this.sendChan:
			if msg.timeout.Before(time.Now()) {
				msg.result <- ErrSendTimeout
			} else if _, err := this.conn.Write(msg.msg); err != nil {
				msg.result <- err
			} else {
				msg.result <- nil
			}
		}
	}
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
			if msg, err := this.decoderMessage(this.conn); err != nil {
				return
			} else if msg == nil {
				//read deadline
			} else {
				this.resetReadIdle()
				this.handlerMessage(msg)
			}
		}
	}
}

func (this *defChannel) handlerMessage(msg interface{}) {
	defer func() {
		if e := recover(); e != nil {
			err := commons.Catch(e)
			if !this.isClosed(err) {
				this.handler.OnError(this, err, msg)
			}
		}
	}()
	this.handler.OnMessage(this, msg)
}

func (this *defChannel) decoderMessage(conn *net.TCPConn) (msg interface{}, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = commons.Catch(e)
			if !this.isClosed(err) {
				this.handler.OnError(this, ErrDecoder, err)
			}
		}
	}()

	_ = this.conn.SetReadDeadline(time.Now().Add(time.Second))
	msg, err = this.coder.Decode(this, conn)
	if err.Error() == "A zero value for t means I/O operations will not time out." {
		err = nil
	}
	return
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
				logrus.Debug("send idle ", this.RemoteAddr(), ", time:", this.idleTimeout)
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

func NewChannel(conn *net.TCPConn, config *RemotingConfig) *defChannel {
	_ = conn.SetNoDelay(true)
	_ = conn.SetKeepAlive(true)
	_ = conn.SetKeepAlivePeriod(time.Duration(config.IdleTime) * time.Second) //这个地方依赖系统

	channel := &defChannel{
		config:     config,
		conn:       conn,
		addr:       conn.RemoteAddr().String(),
		attributes: make(map[string]string),
		closeChan:  make(chan struct{}),
		closeOnce:  &sync.Once{},
		sendChan:   make(chan sendMessage, config.SendLimit),
	}
	channel.idleTimer = time.NewTimer(time.Millisecond * time.Duration(config.IdleTime))
	channel.idleTimeout = 0
	return channel
}
