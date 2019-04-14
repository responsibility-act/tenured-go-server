package client

import (
	"errors"
	"github.com/ihaiker/tenured-go-server/api"
	"github.com/ihaiker/tenured-go-server/commons"
	"github.com/ihaiker/tenured-go-server/commons/registry"
	"github.com/ihaiker/tenured-go-server/commons/snowflake"
	"github.com/ihaiker/tenured-go-server/plugins"
	"github.com/stretchr/testify/assert"
	"testing"
)

var reg registry.ServiceRegistry
var server *AccountServiceClient

func Init() error {
	if plugins, err := plugins.GetRegistryPlugins("consul://127.0.0.1:8500"); err != nil {
		return errors.New("no registry")
	} else {
		if reg, err = plugins.Registry(); err != nil {
			return err
		}
	}
	server, _ = NewAccountServiceClient("tenured_store", reg)

	return server.Start()
}

func Destory() {
	commons.ShutdownIfService(reg, true)
	server.Shutdown(true)
}

func TestNewAccount_Apply(t *testing.T) {
	err := Init()
	assert.Nil(t, err)
	defer Destory()

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
	err := Init()
	assert.Nil(t, err)
	defer Destory()

	ac, err := server.Get(29416244180269568)
	assert.NotNil(t, err)
	t.Log("err=", err)
	t.Log("ac=", ac)
}

func TestAccountServiceClient_Search(t *testing.T) {
	err := Init()
	assert.Nil(t, err)
	defer Destory()

	gl := &registry.GlobalLoading{}
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
