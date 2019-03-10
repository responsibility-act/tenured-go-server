package future

import "github.com/ihaiker/tenured-go-server/commons"

type AsyncRunFuture struct {
	SetFuture
}
type RunableFn func(*AsyncRunFuture) (interface{}, error)

func (self *AsyncRunFuture) run(fn RunableFn) {
	defer func() {
		if e := recover(); e != nil {
			_ = self.Exception(commons.Catch(e))
		}
	}()
	if r, e := fn(self); e != nil {
		_ = self.Exception(e)
	} else {
		_ = self.Set(r)
	}
}

func Run(fn RunableFn) *AsyncRunFuture {
	f := &AsyncRunFuture{}
	f.resultChan = make(chan interface{}, 0)
	f.status = S_RUNNING
	go f.run(fn)
	return f
}
