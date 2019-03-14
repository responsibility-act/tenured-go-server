package executors

import (
	"errors"
	"github.com/ihaiker/tenured-go-server/commons/future"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var fix = NewFixedExecutorService(3, 100)

func print(t *testing.T, i int) func() {
	return func() {
		time.Sleep(time.Millisecond * 20)
		t.Log("out == ", i)
	}
}

func test(t *testing.T, interrupt bool) {
	for i := 0; i < 20; i++ {
		t.Log("put ", i)
		err := fix.Execute(print(t, i))
		if err != nil {
			t.Log(err)
		}
	}
	fix.Shutdown(interrupt)

	err := fix.Execute(print(t, 99))
	assert.NotNil(t, err)
}

func TestFixedExecutorService_Shutdown(t *testing.T) {
	test(t, false)
}

func TestFixedExecutorService_ShutdownNow(t *testing.T) {
	test(t, true)
}

func TestFixedExecutorService_Submit(t *testing.T) {
	f := fix.Submit(func() interface{} {
		return 1
	})
	out, err := f.Get()
	assert.Nil(t, err)
	assert.Equal(t, out, 1)
}

func TestFixedExecutorService_SubmitError(t *testing.T) {
	f := fix.Submit(func() interface{} {
		panic(errors.New("assert"))
		return 1
	})
	out, err := f.Get()
	assert.Equal(t, err.Error(), "assert")
	assert.Nil(t, out)
}

func TestFixedExecutorService_SubmitTimeout(t *testing.T) {
	f := fix.Submit(func() interface{} {
		time.Sleep(time.Second)
		return 1
	})
	out, err := f.GetWithTimeout(time.Millisecond * 10)
	assert.Equal(t, err, future.ErrTimeout)
	assert.Nil(t, out)
}
