package remoting

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type TestHandler struct {
	HandlerWrapper
}

func (h *TestHandler) OnMessage(c RemotingChannel, msg interface{}) {
	logger.Infof("OnMessage %s : msg:%v", c.RemoteAddr(), string(msg.([]byte)))
}

var client = NewRemotingClient(nil)

func init() {
	client.SetHandler(&TestHandler{})
	client.SetCoder(DefaultCoder())
	_ = client.Start()
}

func TestRemotingClient_SendTo(t *testing.T) {
	err := client.SendTo("renzhen.la:9999", []byte("ni hao a"), time.Second*10)
	if err == nil {
		client.Shutdown(true)
	} else {
		t.Log(err)
	}
}

func TestRemotingClient_SyncSendTo(t *testing.T) {
	out := make(chan error, 1)
	client.SyncSendTo("renzhen.la:9999", []byte("ni hao a"), time.Second*10, func(e error) {
		out <- e
	})
	err := <-out
	assert.Nil(t, err)
	client.Shutdown(true)
}

func TestRemotingClient_SendTo_Error(t *testing.T) {
	err := client.SendTo("renzhen.la:9090", []byte("ni hao a"), time.Second)
	assert.NotNil(t, err)
	t.Log(err)
	client.Shutdown(true)
}
