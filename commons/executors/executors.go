package executors

import (
	"github.com/ihaiker/tenured-go-server/commons"
	"github.com/ihaiker/tenured-go-server/commons/future"
)

const (
	ErrShutdown = commons.Error("executor service shutdown")
)

type ExecutorService interface {
	Execute(fn func()) error

	Submit(fn func() interface{}) future.Future

	InvokeAll(fn ...func() interface{}) []future.Future

	Shutdown(interrupt bool)
}
