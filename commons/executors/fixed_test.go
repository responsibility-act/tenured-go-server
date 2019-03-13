package executors

import (
	"testing"
	"time"
)

var fix = NewFixedExecutorService(3, 100)

func print(t *testing.T, i int) func() {
	return func() {
		time.Sleep(time.Second)
		t.Log("out == ", i)
	}
}

func TestFixedExecutorService_Execute(t *testing.T) {
	for i := 0; i < 20; i++ {
		t.Log("put ", i)
		err := fix.Execute(print(t, i))
		if err != nil {
			t.Log(err)
		}
	}
	fix.Shutdown(false)

	err := fix.Execute(print(t, 99))
	if err != nil {
		t.Log(err)
	}
}

func TestFixedExecutorService_ShutdownNow(t *testing.T) {
	for i := 0; i < 20; i++ {
		t.Log("put ", i)
		err := fix.Execute(print(t, i))
		if err != nil {
			t.Log(err)
		}
	}
	fix.Shutdown(true)

	err := fix.Execute(print(t, 99))
	if err != nil {
		t.Log(err)
	}
}
