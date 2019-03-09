package remoting

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewRemotingServer(t *testing.T) {
	err, server := NewRemotingServer(nil)
	assert.Nil(t, err)

	server.SetCoderFactory(func(channel Channel) Coder {
		return DefaultCoder()
	})
	server.SetHandlerFactory(func(channel Channel) Handler {
		return &HandlerWrapper{}
	})
	err = server.Start()
	assert.Nil(t, err)

	time.Sleep(time.Second)

	server.Shutdown()
}
