package ctl

import "github.com/ihaiker/tenured-go-server/protocol"

var (
	errInvalidAccount = protocol.NewError("1000", "Invalid account, authentication failed.")
)
