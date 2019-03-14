package atomic

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestAtomicInt_IncrementAndGet(t *testing.T) {
	a := NewUint32(0)
	wait := sync.WaitGroup{}
	loopFn := func() {
		defer wait.Done()
		for i := 0; i < 1000; i++ {
			a.IncrementAndGet()
		}
	}
	size := 4
	for i := 0; i < size; i++ {
		wait.Add(1)
		go loopFn()
	}
	wait.Wait()
	assert.Equal(t, a.Get(), uint32(1000*size))
}
