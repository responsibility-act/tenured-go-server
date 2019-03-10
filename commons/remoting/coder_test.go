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
	bs, err := coder.Encode(1)
	assert.Equal(t, err, os.ErrInvalid)

	bs, err = coder.Encode([]byte("test"))
	assert.Nil(t, err)
	assert.Equal(t, bs, []byte("test"))
}
