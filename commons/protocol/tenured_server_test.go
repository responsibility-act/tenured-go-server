package protocol

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestTenuredServer(t *testing.T) {
	server, err := NewTenuredServer(":6071", nil)
	assert.Nil(t, err)

	err = server.Start()
	assert.Nil(t, err)

	time.Sleep(time.Hour)

	server.Shutdown(true)
}
