package protocol

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const HELLO = uint16(2)

func TestTenuredClient(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)

	client, err := NewTenuredClient(nil)
	client.Module = &AuthHeader{
		Module:     "test",
		Address:    "127.0.0.1:8080",
		Attributes: map[string]string{"testkey": "testvalue"},
	}
	assert.Nil(t, err)

	err = client.Start()
	assert.Nil(t, err)

	request := NewRequest(HELLO)
	request.Body = []byte("hello")
	response, err := client.Invoke("127.0.0.1:6071", request, time.Second)
	assert.Nil(t, err)

	if !response.IsSuccess() {
		err := response.GetError().(*TenuredError)
		t.Log("send error: code=", err.Code, " mesage=", err.Message)
	}

	client.Shutdown(true)
}

func TestTenuredServer(t *testing.T) {
	server, err := NewTenuredServer(":6071", nil)
	assert.Nil(t, err)

	err = server.Start()
	assert.Nil(t, err)

	t.Log("start shutdown server")
	server.Shutdown(true)
	t.Log("end shutdown server")
}
