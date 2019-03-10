package remoting

import (
	"io"
	"os"
)

type Coder interface {
	Decode(io.Reader) (interface{}, error)
	Encode(interface{}) ([]byte, error)
}

type CoderFactory func(Channel, RemotingConfig) Coder

type Bytes1024Coder struct{}

func (this *Bytes1024Coder) Decode(reader io.Reader) (interface{}, error) {
	bs := make([]byte, 1024)
	if length, err := reader.Read(bs); err == nil && length > 0 {
		return bs[:length], nil
	} else if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

func (this *Bytes1024Coder) Encode(msg interface{}) ([]byte, error) {
	if bs, ok := msg.([]byte); ok {
		return bs, nil
	} else {
		return nil, os.ErrInvalid
	}
}

func DefaultCoder() Coder {
	return &Bytes1024Coder{}
}
