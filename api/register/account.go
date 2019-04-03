package register

import (
	"github.com/ihaiker/tenured-go-server/api"
	"github.com/ihaiker/tenured-go-server/commons/protocol"
	"github.com/ihaiker/tenured-go-server/commons/registry"
	"time"
)

type accountServerClient struct {
	*protocol.TenuredClientInvoke
}

func (this *accountServerClient) Apply(account *api.Account) (*api.Account, *protocol.TenuredError) {
	respAccount := &api.Account{}
	err := this.Invoke(api.AccountServiceApply, account, nil, time.Second*3, respAccount)
	if err != nil {
		respAccount = nil
	}
	return respAccount, err
}

func NewAccountServer(serverName string, reg registry.ServiceRegistry) (api.AccountService, error) {
	loadBalance := registry.NewRangeLoadBalance()
	client := &accountServerClient{
		TenuredClientInvoke: protocol.NewClient(serverName, reg, loadBalance),
	}
	err := client.Start()
	return client, err
}
