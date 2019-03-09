package remoting

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var coder = DefaultCoder()

func TestDefaultCoder_Decode(t *testing.T) {
	reader := bytes.NewReader([]byte("test"))
	bs, err := coder.Decode(reader)
	assert.Nil(t, err)
	assert.Equal(t, "test", string(bs.([]byte)))
}

func TestBytes1024Coder_Encode(t *testing.T) {
	writer := bytes.NewBuffer(make([]byte, 0))
	err := coder.Encode(1, writer)
	assert.Equal(t, err, os.ErrInvalid)

	err = coder.Encode([]byte("test"), writer)
	assert.Nil(t, err)
	bs := make([]byte, 1024)
	r, err := writer.Read(bs)
	assert.Nil(t, err)
	assert.Equal(t, r, 4)
	assert.Equal(t, "test", string(bs[:r]))
}
