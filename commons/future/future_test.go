package future

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var f = Set()

func init() {
	go func() {
		time.Sleep(time.Second)
		b := f.Set("out")
		println("set out: ", b)
	}()
}

func TestFuture_Get(t *testing.T) {
	out, err := f.Get()
	assert.Nil(t, err)
	assert.Equal(t, out, "out")
}

func TestFuture_GetWithTimeout(t *testing.T) {
	out, err := f.GetWithTimeout(time.Microsecond * 200)
	assert.Equal(t, err, ErrTimeout)
	assert.Nil(t, out)

	out, err = f.GetWithTimeout(time.Microsecond * 200)
	assert.Equal(t, err, ErrTimeout)
	assert.Nil(t, out)

	f.Cancel()

	out, err = f.Get()
	assert.Equal(t, err, ErrFutureCancel)
	assert.Nil(t, out)
}

func TestFuture_GetOver(t *testing.T) {
	out, err := f.GetWithTimeout(time.Microsecond * 200)
	assert.Equal(t, err, ErrTimeout)
	assert.Nil(t, out)

	out, err = f.Get()
	assert.Nil(t, err)
	assert.Equal(t, out, "out")
}
