package remoting

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewRemotingServer(t *testing.T) {
	server, err := NewRemotingServer(":6071", nil)
	assert.Nil(t, err)

	server.SetCoderFactory(func(channel RemotingChannel, config RemotingConfig) RemotingCoder {
		return DefaultCoder()
	})
	server.SetHandlerFactory(func(channel RemotingChannel, config RemotingConfig) RemotingHandler {
		return &HandlerWrapper{}
	})
	err = server.Start()
	assert.Nil(t, err)

	err = server.SendTo("127.0.0.1:8080", []byte("123123"), time.Second)
	t.Log("sendto 127.0.0.1:8080", err)

	server.Shutdown(true)
}
