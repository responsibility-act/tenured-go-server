package account

import (
	"github.com/ihaiker/tenured-go-server/api"
	"github.com/ihaiker/tenured-go-server/api/command"
	"github.com/ihaiker/tenured-go-server/commons/protocol"
	"github.com/ihaiker/tenured-go-server/commons/registry"
	"time"
)

type accountServerClient struct {
	*protocol.ServerClient
}

func (this *accountServerClient) Apply(account *command.Account) (*command.Account, *protocol.TenuredError) {
	respAccount := &command.Account{}
	err := this.Invoke(api.AccountServiceApply, account, nil, time.Second*3, respAccount)
	return respAccount, err
}

func NewAccountServer(serverName string, reg registry.ServiceRegistry) (api.AccountService, error) {
	loadBalance := registry.NewRangeLoadBalance()
	client := &accountServerClient{
		ServerClient: protocol.NewClient(serverName, reg, loadBalance),
	}
	err := client.Start()
	return client, err
}
