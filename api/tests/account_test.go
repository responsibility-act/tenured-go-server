package tests

import (
	"github.com/ihaiker/tenured-go-server/api"
	"github.com/ihaiker/tenured-go-server/api/client"
	"github.com/ihaiker/tenured-go-server/commons/snowflake"
	"github.com/ihaiker/tenured-go-server/registry"
	"github.com/ihaiker/tenured-go-server/registry/load_balance"
	"github.com/ihaiker/tenured-go-server/registry/plugins"
	"github.com/stretchr/testify/assert"
	"testing"
)

func GetAccountService() (server *client.AccountServiceClient, reg registry.ServiceRegistry, err error) {
	var plugin registry.Plugins
	if plugin, err = plugins.GetRegistryPlugins("consul://127.0.0.1:8500"); err != nil {
		return
	} else {
		if reg, err = plugin.Registry(); err != nil {
			return
		}
	}
	server = client.NewAccountServiceClient(load_balance.NewRoundLoadBalance("tenured_store", api.StoreAccount, reg))

	if err = server.Start(); err != nil {
		return
	}
	return
}

func TestNewAccount_Apply(t *testing.T) {
	server, _, err := GetAccountService()
	assert.Nil(t, err)

	id, _ := snowflake.NewSnowflake(snowflake.Settings{}).NextID()

	account := &api.Account{}
	account.Id = id
	account.Email = "wo@renzhen.la"

	if err := server.Apply(account); err != nil {
		t.Log(err)
	}

	ac, err := server.Get(account.Id)
	assert.Nil(t, err)

	t.Log(ac)
}

func TestAccountService_Get(t *testing.T) {
	server, _, err := GetAccountService()
	assert.Nil(t, err)

	ac, err := server.Get(29416244180269568)
	assert.NotNil(t, err)
	t.Log("err=", err)
	t.Log("ac=", ac)
}

func TestAccountServiceClient_Search(t *testing.T) {
	server, _, err := GetAccountService()
	assert.Nil(t, err)

	gl := &load_balance.GlobalLoading{}
	search := new(api.Search)
	search.Limit = 10
	for gl.NextNode() {
		rs, err := server.Search(gl, search)
		assert.Nil(t, err)
		t.Log("Search In: ", gl.Server.Id)
		for _, a := range rs.Accounts {
			t.Log(a)
		}
	}

}
