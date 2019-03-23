package remoting

import (
	"errors"
	"github.com/ihaiker/tenured-go-server/commons"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type RemotingChannel interface {
	RemoteAddr() string

	Attributes() map[string]interface{}

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

	attributes map[string]interface{}

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
func (this *defChannel) Attributes() map[string]interface{} {
	return this.attributes
}

func (this *defChannel) encodeMessage(msg interface{}) (bs []byte, err error) {
	defer func() {
		if e := recover(); e != nil {
			catchErr := commons.Catch(e)
			if !this.isClosed(catchErr) {
				err = &RemotingError{Op: ErrEncoder, Err: catchErr}
				this.handler.OnError(this, err, msg)
			}
		}
	}()

	if bs, err = this.coder.Encode(this, msg); err != nil {
		if rerr, ok := err.(*RemotingError); ok {
			err = rerr
		} else {
			err = &RemotingError{Op: ErrDecoder, Err: err}
		}
		return nil, err
	} else if len(bs) > this.config.PacketBytesLimit {
		return nil, &RemotingError{Op: ErrPacketBytesLimit, Err: errors.New("the packet limit size " + strconv.Itoa(this.config.PacketBytesLimit))}
	} else {
		return bs, nil
	}
}

func (this *defChannel) write(msg interface{}, timeout time.Duration, callback func(error)) error {
	timeoutTime := time.Now().Add(timeout)
	if bs, err := this.encodeMessage(msg); err != nil {
		if callback != nil {
			callback(err)
		}
		return err
	} else {
		if timeoutTime.Before(time.Now()) { //encode timeout
			err := &RemotingError{Op: ErrSendTimeout, Err: errors.New("send timeout: " + timeout.String())}
			if callback != nil {
				callback(err)
			}
			return err
		}
		fn := func() error {
			result := make(chan error, 1)
			this.sendChan <- sendMessage{msg: bs, timeout: time.Now().Add(timeout), result: result}
			err := <-result
			close(result)
			if callback != nil {
				callback(err)
			}
			return err
		}
		if callback == nil {
			return fn()
		} else {
			go fn()
			return nil
		}
	}
}

func (this *defChannel) Write(msg interface{}, timeout time.Duration) error {
	return this.write(msg, timeout, nil)
}
func (this *defChannel) AsyncWrite(msg interface{}, timeout time.Duration, callback func(error)) {
	_ = this.write(msg, timeout, callback)
}

func (this *defChannel) Do(onClose func(channel RemotingChannel)) error {
	this.onCloseFn = onClose
	go this.syncDo(this.readLoop)
	if this.config.IdleTime > 0 {
		go this.syncDo(this.heartbeatLoop)
	}
	go this.syncDo(this.writeLoop)
	err := this.handler.OnChannel(this)
	if err != nil {
		this.Close()
	}
	return err
}

func (this *defChannel) Close() {
	this.closeOnce.Do(func() {
		logrus.Infof("close channel: %s", this.RemoteAddr())
		this.idleTimer.Stop()
		this.handler.OnClose(this)
		if this.onCloseFn != nil {
			this.onCloseFn(this)
		}
		_ = this.conn.Close()
		close(this.closeChan)
	})
}

func (this *defChannel) closeUnWriteMessageChan() {
	for {
		select {
		case <-time.After(time.Second):
			return
		case msg := <-this.sendChan:
			msg.result <- &RemotingError{Op: ErrClosed, Err: errors.New("the channel is closed")}
		}
	}
}

func (this *defChannel) writeLoop() {
	defer func() {
		this.closeUnWriteMessageChan()
		this.Close()
	}()
	logrus.Debug("start write loop:", this.RemoteAddr())

	for {
		select {
		case <-this.closeChan:
			return
		case msg := <-this.sendChan:
			if msg.timeout.Before(time.Now()) {
				msg.result <- &RemotingError{Op: ErrSendTimeout, Err: errors.New("send timeout")}
			} else if _, err := this.conn.Write(msg.msg); err != nil {
				msg.result <- err
			} else {
				msg.result <- nil
			}
		}
	}
}

func (this *defChannel) readLoop() {
	defer this.Close()
	logrus.Debug("start read loop:", this.RemoteAddr())
	for {
		select {
		case <-this.closeChan:
			return
		default:
			if msg, err := this.decoderMessage(this.conn); err != nil {
				logrus.Infof("decode error close channel %s, error:%s ", this.RemoteAddr(), err)
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
				this.handler.OnError(this, &RemotingError{Op: ErrDecoder, Err: err}, nil)
			}
		}
	}()

	_ = this.conn.SetReadDeadline(time.Now().Add(time.Second))
	msg, err = this.coder.Decode(this, conn)
	if err != nil {
		if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
			err = nil
		}
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

func (this *defChannel) heartbeatTimeout() time.Duration {
	return time.Second * time.Duration(this.config.IdleTime)
}
func (this *defChannel) resetReadIdle() {
	this.idleTimer.Reset(this.heartbeatTimeout())
	this.idleTimeout = 0
}

func (this *defChannel) heartbeatLoop() {
	logrus.Debug("start heartbeat loop: ", this.RemoteAddr())
	defer this.Close()

	idleCheckTime := this.heartbeatTimeout()
	for {
		select {
		case <-this.closeChan:
			return
		case t := <-this.idleTimer.C:
			timestr := t.Format("2006-01-02 15:04:05")
			if this.idleTimeout+1 <= this.config.IdleTimeout {
				logrus.Infof("SendIdle to: %s, time: %s", this.RemoteAddr(), timestr)
				this.idleTimeout = this.idleTimeout + 1
				this.idleTimer.Reset(idleCheckTime)
				this.handler.OnIdle(this)
			} else {
				logrus.Infof("IdleTimerOut: %s, %s", this.RemoteAddr(), timestr)
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
		attributes: map[string]interface{}{},
		closeChan:  make(chan struct{}),
		closeOnce:  &sync.Once{},
		sendChan:   make(chan sendMessage, config.SendLimit),
	}
	channel.idleTimer = time.NewTimer(channel.heartbeatTimeout())
	channel.idleTimeout = 0
	return channel
}
