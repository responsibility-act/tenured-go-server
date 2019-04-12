package executors

import "github.com/ihaiker/tenured-go-server/commons/future"

type queueBucket struct {
	fn func() interface{}
	fu *future.SetFuture
}
