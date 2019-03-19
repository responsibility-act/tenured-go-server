package protocol

import (
	"bytes"
	"github.com/ihaiker/tenured-go-server/commons/remoting"
	"github.com/stretchr/testify/assert"
	"testing"
)

var c = tenuredCoder{config: remoting.DefaultConfig()}

func TestTenuredCoder_Request(t *testing.T) {
	request := NewRequest(1)
	_ = request.SetHeader(map[string]string{"name": "value"})
	request.Body = []byte("testbody")
	bs, err := c.Encode(nil, request)
	assert.Nil(t, err)

	t.Log(string(bs))

	reader := bytes.NewReader(bs)
	d1, err := c.Decode(nil, reader)
	assert.Nil(t, err)

	decodeReq := d1.(*TenuredCommand)
	assert.Equal(t, decodeReq.id, request.id)
	assert.Equal(t, decodeReq.id, uint32(1))
	assert.Equal(t, decodeReq.code, request.code)
	assert.Equal(t, decodeReq.code, uint16(1))
	assert.Equal(t, decodeReq.flag, request.flag)
	assert.Equal(t, decodeReq.header, request.header)
	assert.Equal(t, string(decodeReq.header), `{"name":"value"}`)
	assert.Equal(t, decodeReq.Body, request.Body)
	assert.Equal(t, decodeReq.Body, []byte("testbody"))
}

func TestTenuredCoder_Response(t *testing.T) {
	response := NewACK(1)
	response.RemotingError(ErrorNoAuth())

	bs, err := c.Encode(nil, response)

	reader := bytes.NewReader(bs)
	d1, err := c.Decode(nil, reader)
	assert.Nil(t, err)

	dResp := d1.(*TenuredCommand)
	assert.False(t, dResp.IsSuccess())
	t.Log(string(dResp.header))
	t.Log(string(dResp.Body))
}
