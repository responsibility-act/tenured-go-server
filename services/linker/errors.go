package linker

import "github.com/ihaiker/tenured-go-server/protocol"

var (
	ErrAuth = protocol.NewError("1001", "Authentication failed！")
)
