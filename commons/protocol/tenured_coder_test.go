package protocol

import (
	"bytes"
	"github.com/ihaiker/tenured-go-server/commons/remoting"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTenuredCoder(t *testing.T) {
	c := tenuredCoder{config: remoting.DefaultConfig()}

	request := NewRequest(1)
	_ = request.SetHeader(map[string]string{"name": "value"})
	request.Body = []byte("testbody")
	bs, err := c.Encode(nil, request)
	assert.Nil(t, err)

	reader := bytes.NewReader(bs)
	d1, err := c.Decode(nil, reader)
	assert.Nil(t, err)

	decodeReq := d1.(*TenuredCommand)
	assert.Equal(t, decodeReq.Id, request.Id)
	assert.Equal(t, decodeReq.Id, uint32(1))
	assert.Equal(t, decodeReq.Code, request.Code)
	assert.Equal(t, decodeReq.Code, uint16(1))
	assert.Equal(t, decodeReq.Flag, request.Flag)
	assert.Equal(t, decodeReq.Header, request.Header)
	assert.Equal(t, string(decodeReq.Header), `{"name":"value"}`)
	assert.Equal(t, decodeReq.Body, request.Body)
	assert.Equal(t, decodeReq.Body, []byte("testbody"))
}
