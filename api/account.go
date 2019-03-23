package api

import (
	"github.com/ihaiker/tenured-go-server/api/command"
	"github.com/ihaiker/tenured-go-server/commons/protocol"
)

type AccountService interface {
	//申请一个账户
	Apply(applyAccount *command.Account) (account *command.Account, err *protocol.TenuredError)
}
