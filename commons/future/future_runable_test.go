package future

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var fr = Run(func(future *AsyncRunFuture) (i interface{}, e error) {
	time.Sleep(time.Second)
	return "runout", nil
})

func TestRunFuture_Get(t *testing.T) {
	out, err := fr.Get()
	assert.Nil(t, err)
	assert.Equal(t, out, "runout")
}

func TestRunFuture_GetWithTimeout(t *testing.T) {
	out, err := fr.GetWithTimeout(time.Microsecond * 200)
	assert.Equal(t, err, ErrTimeout)
	assert.Nil(t, out)

	out, err = fr.GetWithTimeout(time.Microsecond * 200)
	assert.Equal(t, err, ErrTimeout)
	assert.Nil(t, out)

	fr.Cancel()

	out, err = fr.Get()
	assert.Equal(t, err, ErrFutureCancel)
	assert.Nil(t, out)
}

func TestRunFuture_GetOver(t *testing.T) {
	out, err := fr.GetWithTimeout(time.Microsecond * 200)
	assert.Equal(t, err, ErrTimeout)
	assert.Nil(t, out)

	out, err = fr.Get()
	assert.Nil(t, err)
	assert.Equal(t, out, "runout")
}
