package protocol

import (
	"github.com/ihaiker/tenured-go-server/commons/executors"
	"github.com/ihaiker/tenured-go-server/commons/remoting"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const HELLO = uint16(2)
const HEADER = uint16(3)

var server *TenuredServer
var client *TenuredClient

func startServer() *TenuredServer {
	config := remoting.DefaultConfig()
	config.IdleTime = 1
	server, _ := NewTenuredServer(":6071", config)
	server.AuthHeader = &AuthHeader{
		Module:     "test",
		Address:    "127.0.0.1:6071",
		Attributes: map[string]string{},
	}

	executorService := executors.NewFixedExecutorService(1, 10)

	server.RegisterCommandProcesser(HELLO, func(channel remoting.RemotingChannel, command *TenuredCommand) {
		ack := NewACK(command.ID())
		ack.Body = []byte(string(command.Body) + " tenured")
		_ = channel.Write(ack, time.Second)
	}, executorService)

	server.RegisterCommandProcesser(HEADER, func(channel remoting.RemotingChannel, command *TenuredCommand) {
		ack := NewACK(command.ID())
		if err := ack.SetHeader(map[string]string{"hello": "tenured"}); err != nil {
			logger.Error(err)
		} else {
			_ = channel.Write(ack, time.Second)
		}

	}, executorService)

	_ = server.Start()
	return server
}

func startCleint() *TenuredClient {
	clientConfig := remoting.DefaultConfig()
	clientConfig.IdleTime = 0
	client, _ := NewTenuredClient(clientConfig)
	client.AuthHeader = &AuthHeader{
		Module:     "test",
		Address:    "127.0.0.1:8080",
		Attributes: map[string]string{"testkey": "testvalue"},
	}
	_ = client.Start()
	return client
}

func init() {
	server = startServer()
	client = startCleint()
}

func destory() {
	client.Shutdown(true)
	server.Shutdown(true)
}

func TestTenured_hello(t *testing.T) {
	request := NewRequest(HELLO)
	request.Body = []byte("hello")
	response, err := client.Invoke("127.0.0.1:6071", request, time.Second)
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.True(t, response.IsSuccess())
	assert.Equal(t, "hello tenured", string(response.Body))

	header := map[string]string{}
	err = response.GetHeader(header)
	assert.Equal(t, err, ErrNoHeader)

	destory()
}

func TestTenured_Header(t *testing.T) {
	request := NewRequest(HEADER)
	request.Body = []byte("hello")
	response, err := client.Invoke("127.0.0.1:6071", request, time.Second)
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.True(t, response.IsSuccess())

	header := map[string]string{}
	err = response.GetHeader(&header)
	assert.Nil(t, err)
	assert.Equal(t, "tenured", header["hello"])
	destory()
}
