package future

type SetFuture struct {
	futureWapper
}

func (self *SetFuture) Set(result interface{}) bool {
	if self.status != S_RUNNING {
		return false
	}
	if self.atomicSet(S_RUNNING, S_OVER) {
		self.result = result
		close(self.resultChan)
		return true
	}
	return false
}
func (self *SetFuture) Exception(err error) bool {
	if self.status != S_RUNNING {
		return false
	}
	if self.atomicSet(S_RUNNING, S_EXCEPTION) {
		self.err = err
		close(self.resultChan)
		return true
	}
	return false
}

func Set() *SetFuture {
	f := &SetFuture{}
	f.resultChan = make(chan interface{}, 1)
	f.status = S_RUNNING
	return f
}
