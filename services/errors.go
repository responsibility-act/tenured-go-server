package services

import "github.com/ihaiker/tenured-go-server/protocol"

var (
	ErrInvalidJson    = protocol.NewError("1000", "Invalid Body Json")
	ErrInvalidUserId  = protocol.NewError("1001", "Invalid UserId")
	ErrInvalidAccount = protocol.NewError("1000", "Invalid account, authentication failed.")
)
