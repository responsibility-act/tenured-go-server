package commons

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestServerStatus(t *testing.T) {
	status := S_STATUS_INIT

	b := status.Shutdown(func() {
		t.Log("no do...")
	})
	assert.False(t, b)

	status.Start(func() {
		t.Log("start ..")
	})

	status.ReStart(func() {
		t.Log("restart...")
	})
	assert.Equal(t, status, S_STATUS_UP)

	t.Log(status)
}
