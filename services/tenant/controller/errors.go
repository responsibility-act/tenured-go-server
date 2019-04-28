package ctl

import "github.com/ihaiker/tenured-go-server/protocol"

var (
	errInvoildAccount = protocol.NewError("1000", "Invalid account, authentication failed.")
)
