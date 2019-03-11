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
	bs, err := coder.Decode(nil, reader)
	assert.Nil(t, err)
	assert.Equal(t, "test", string(bs.([]byte)))
}

func TestBytes1024Coder_Encode(t *testing.T) {
	bs, err := coder.Encode(nil, 1)
	assert.Equal(t, err, os.ErrInvalid)

	bs, err = coder.Encode(nil, []byte("test"))
	assert.Nil(t, err)
	assert.Equal(t, bs, []byte("test"))
}
