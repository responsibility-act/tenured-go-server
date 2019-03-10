package future

import (
	"github.com/ihaiker/tenured-go-server/commons"
	"sync/atomic"
	"time"
)

const (
	S_RUNNING uint32 = iota
	S_CANCEL
	S_EXCEPTION
	S_OVER
)
const (
	ErrFutureCancel = commons.Error("cancel")
	ErrTimeout      = commons.Error("timeout")
)

type Future interface {
	Cancel() bool

	IsCancelled() bool

	IsDone() bool

	Get() (interface{}, error)

	GetWithTimeout(timeout time.Duration) (interface{}, error)
}

type futureWapper struct {
	result     interface{}
	err        error
	resultChan chan interface{}
	status     uint32
}

func (self *futureWapper) atomicSet(old, new uint32) bool {
	return atomic.CompareAndSwapUint32(&self.status, old, new)
}

func (self *futureWapper) atomicGet() uint32 {
	return atomic.LoadUint32(&self.status)
}

func (self *futureWapper) is(check ...uint32) bool {
	for _, v := range check {
		if atomic.LoadUint32(&self.status) == v {
			return true
		}
	}
	return false
}

func (self *futureWapper) Cancel() bool {
	if self.status != S_RUNNING {
		return false
	}
	if self.atomicSet(S_RUNNING, S_CANCEL) {
		self.err = ErrFutureCancel
		close(self.resultChan)
		return true
	}
	return false
}

func (self *futureWapper) IsCancelled() bool {
	return self.is(S_CANCEL)
}

func (self *futureWapper) IsDone() bool {
	return self.is(S_EXCEPTION, S_OVER, S_CANCEL)
}

func (self *futureWapper) Get() (interface{}, error) {
	if self.IsDone() {
		return self.result, self.err
	}
	<-self.resultChan
	return self.result, self.err
}

func (self *futureWapper) GetWithTimeout(timeout time.Duration) (interface{}, error) {
	if self.IsDone() {
		return self.result, self.err
	}
	select {
	case <-self.resultChan:
		return self.result, self.err
	case <-time.After(timeout):
		return nil, ErrTimeout
	}
}
