package remoting

import (
	"fmt"
)

type ErrorType string

func (this ErrorType) String() string {
	return string(this)
}

const (
	ErrCoder   = ErrorType("Coder")
	ErrEncoder = ErrorType("Encoder")
	ErrDecoder = ErrorType("Decoder")

	ErrHandler = ErrorType("Handler")

	ErrPacketBytesLimit = ErrorType("PacketBytesLimit")
	ErrClosed           = ErrorType("Closed")
	ErrSendTimeout      = ErrorType("Timeout")

	ErrNoChannel = ErrorType("NoChannel")
)

type RemotingError struct {
	Err error
	Op  ErrorType
}

func (this *RemotingError) Error() string {
	return fmt.Sprintf("[%s]%s", this.Op, this.Err.Error())
}

func IsRemotingError(err error, types ...ErrorType) bool {
	if e, is := err.(*RemotingError); is {
		for _, t := range types {
			if t == e.Op {
				return true
			}
		}
	}
	return false
}
