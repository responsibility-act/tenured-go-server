package commons

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTry(t *testing.T) {
	Try(func() {
		a := 0
		b := 2
		c := b / a
		t.Log(c)
	}, func(e error) {
		t.Log(e)
		assert.NotNil(t, e)
	})
}

func TestPanic(t *testing.T) {
	Try(func() {
		panic("new")
	}, func(e error) {
		t.Log(e)
		assert.NotNil(t, e)
	})
}
