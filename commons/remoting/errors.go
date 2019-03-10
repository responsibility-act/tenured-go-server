package remoting

import "github.com/ihaiker/tenured-go-server/commons"

const (
	ErrNoCoder          = commons.Error("coder is nil")
	ErrNoHandler        = commons.Error("handler is nil")
	ErrPacketBytesLimit = commons.Error("packet is limit")
	ErrEncoder          = commons.Error("error in encoder")
	ErrDecoder          = commons.Error("error in decoder")
)
