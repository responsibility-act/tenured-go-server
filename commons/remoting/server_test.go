package remoting

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewRemotingServer(t *testing.T) {
	server, err := NewRemotingServer(nil)
	assert.Nil(t, err)

	server.SetCoderFactory(func(channel Channel, config RemotingConfig) Coder {
		return DefaultCoder()
	})
	server.SetHandlerFactory(func(channel Channel, config RemotingConfig) Handler {
		return &HandlerWrapper{}
	})
	err = server.Start()
	assert.Nil(t, err)

	time.Sleep(time.Second)

	server.Shutdown()
}
