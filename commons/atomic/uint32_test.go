package atomic

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestAtomicInt_IncrementAndGet(t *testing.T) {
	a := AtomicUInt32{value: 0}

	loopFn := func() {
		defer println("OVER ...")
		for i := 0; i < 1000; i++ {
			a.IncrementAndGet()
		}
	}
	go loopFn()
	go loopFn()

	time.Sleep(time.Second)

	assert.Equal(t, a.Get(), uint32(2000))
}
