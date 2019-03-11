package remoting

import (
	"io"
	"os"
)

type RemotingCoder interface {
	Decode(RemotingChannel, io.Reader) (interface{}, error)
	Encode(RemotingChannel, interface{}) ([]byte, error)
}

type RemotingCoderFactory func(RemotingChannel, RemotingConfig) RemotingCoder

type Bytes1024Coder struct{}

func (this *Bytes1024Coder) Decode(channel RemotingChannel, reader io.Reader) (interface{}, error) {
	bs := make([]byte, 1024)
	if length, err := reader.Read(bs); err == nil && length > 0 {
		return bs[:length], nil
	} else if err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

func (this *Bytes1024Coder) Encode(channel RemotingChannel, msg interface{}) ([]byte, error) {
	if bs, ok := msg.([]byte); ok {
		return bs, nil
	} else {
		return nil, os.ErrInvalid
	}
}

func DefaultCoder() RemotingCoder {
	return &Bytes1024Coder{}
}
