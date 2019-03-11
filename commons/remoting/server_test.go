package remoting

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewRemotingServer(t *testing.T) {
	server, err := NewRemotingServer(nil)
	assert.Nil(t, err)

	server.SetCoderFactory(func(channel RemotingChannel, config RemotingConfig) RemotingCoder {
		return DefaultCoder()
	})
	server.SetHandlerFactory(func(channel RemotingChannel, config RemotingConfig) RemotingHandler {
		return &HandlerWrapper{}
	})
	err = server.Start()
	assert.Nil(t, err)

	time.Sleep(time.Second)

	server.Shutdown()
}
