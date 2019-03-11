package remoting

import "github.com/ihaiker/tenured-go-server/commons"

const (
	ErrNoCoder          = commons.Error("coder is nil")
	ErrNoHandler        = commons.Error("handler is nil")
	ErrPacketBytesLimit = commons.Error("packet is limit")
	ErrClosed           = commons.Error("closed")
	ErrSendTimeout      = commons.Error("send timeout")
	ErrEncoder          = commons.Error("error in encoder")
	ErrDecoder          = commons.Error("error in decoder")
)
