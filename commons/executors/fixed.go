package executors

import (
	"github.com/ihaiker/tenured-go-server/commons"
	"github.com/ihaiker/tenured-go-server/commons/future"
	"sync"
)

type queueBucket struct {
	fn func() interface{}
	fu *future.SetFuture
}

type fixedExecutorService struct {
	queue     chan *queueBucket
	waitGroup *sync.WaitGroup
	size      int
	interrupt bool
	closeChan chan struct{} /*interrupt*/
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

func (this *fixedExecutorService) InvokeAll(fn ...func() interface{}) []future.Future {
	all := make([]future.Future, len(fn))
	for i := 0; i < len(fn); i++ {
		all[i] = this.Submit(fn[i])
	}
	return all
}

func (this *fixedExecutorService) Shutdown(interrupt bool) {
	this.status.Shutdown(func() {
		this.interrupt = interrupt
		this.close()
	})
	this.waitGroup.Wait()
}

func (this *fixedExecutorService) close() {
	close(this.closeChan)
	close(this.queue)
}

func (this *fixedExecutorService) start() {
	this.status.Start(func() {
		for i := 0; i < this.size; i++ {
			this.waitGroup.Add(1)
			go this.run(i)
		}
	})
}

func execFn(qb *queueBucket) {
	defer func() {
		if e := recover(); e != nil {
			err := commons.Catch(e)
			if qb.fu != nil {
				qb.fu.Exception(err)
			}
		}
	}()

	if out := qb.fn(); qb.fu != nil {
		qb.fu.Set(out)
	}
}

func (this *fixedExecutorService) run(i int) {
	defer this.waitGroup.Done()

LOOP:
	for {
		select {
		case <-this.closeChan:
			if this.interrupt {
				return
			} else {
				break LOOP
			}
		case queueZone := <-this.queue:
			if queueZone != nil {
				execFn(queueZone)
			}
		}
	}

	for {
		if queueZone := <-this.queue; queueZone == nil {
			return
		} else {
			execFn(queueZone)
		}
	}
}

func NewFixedExecutorService(size int, queueSize int) ExecutorService {
	service := &fixedExecutorService{
		size:      size,
		closeChan: make(chan struct{}),
		waitGroup: &sync.WaitGroup{},
		queue:     make(chan *queueBucket, queueSize),
		status:    commons.S_STATUS_INIT,
	}
	service.start()
	return service
}
