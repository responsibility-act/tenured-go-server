package executors

import (
	"github.com/ihaiker/tenured-go-server/commons"
	at "github.com/ihaiker/tenured-go-server/commons/atomic"
	"github.com/ihaiker/tenured-go-server/commons/future"
	"sync"
	"time"
)

type queueBucket struct {
	fn func() interface{}
	fu *future.SetFuture
}

type fixedExecutorService struct {
	queue     chan *queueBucket
	waitGroup *sync.WaitGroup
	size      *at.AtomicUInt32

	closeChan chan bool
	status    commons.ServerStatus
}

func (this *fixedExecutorService) Execute(fn func()) error {
	if !this.status.IsUp() {
		return ErrShutdown
	}
	this.queue <- &queueBucket{
		fn: func() interface{} {
			fn()
			return nil
		},
	}
	return nil
}

func (this *fixedExecutorService) Submit(fn func() interface{}) future.Future {
	fu := future.Set()
	if !this.status.IsUp() {
		fu.Exception(ErrShutdown)
		return fu
	}
	this.queue <- &queueBucket{fn: fn, fu: fu}
	return fu
}

func (this *fixedExecutorService) InvokeAll(fn ...func() interface{}) []future.SetFuture {
	return nil
}

func (this *fixedExecutorService) Shutdown(interrupt bool) {
	this.status.Shutdown(func() {
		this.closeChan <- interrupt
	})
	this.waitGroup.Wait()
}

func (this *fixedExecutorService) close() {
	close(this.closeChan)
	close(this.queue)
}

func (this *fixedExecutorService) start() {
	this.status.Start(func() {
		for i := 0; i < int(this.size.Get()); i++ {
			this.waitGroup.Add(1)
			go this.run(i)
		}
	})
}

func (this *fixedExecutorService) run(i int) {
	defer func() {
		this.close()
		this.waitGroup.Done()
	}()

LOOP:
	for {
		select {
		case interrupt := <-this.closeChan:
			if interrupt {
				return
			} else {
				break LOOP
			}
		case queueZone := <-this.queue:
			out := queueZone.fn()
			if queueZone.fu != nil {
				queueZone.fu.Set(out)
			}
		}
	}

	for {
		select {
		case <-time.After(time.Millisecond * 500):
			return
		case queueZone := <-this.queue:
			if queueZone == nil {
				return
			}
			out := queueZone.fn()
			if queueZone.fu != nil {
				queueZone.fu.Set(out)
			}
		}
	}
}

func NewFixedExecutorService(size int, queueSize int) ExecutorService {
	service := &fixedExecutorService{
		size:      at.NewUint32(uint32(size)),
		closeChan: make(chan bool),
		waitGroup: &sync.WaitGroup{},
		queue:     make(chan *queueBucket, queueSize),
		status:    commons.S_STATUS_INIT,
	}
	service.start()
	return service
}
