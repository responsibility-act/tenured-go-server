package tests

import (
	"github.com/ihaiker/tenured-go-server/api"
	"github.com/ihaiker/tenured-go-server/api/client"
	"github.com/ihaiker/tenured-go-server/registry"
	"github.com/ihaiker/tenured-go-server/registry/load_balance"
	"github.com/ihaiker/tenured-go-server/registry/plugins"
	"github.com/stretchr/testify/assert"
	"testing"
)

func GetUserService() (server *client.UserServiceClient) {
	var reg registry.ServiceRegistry
	var err error
	var plugin registry.Plugins
	if plugin, err = plugins.GetRegistryPlugins("consul://127.0.0.1:8500"); err != nil {
		return
	} else {
		if reg, err = plugin.Registry(); err != nil {
			return
		}
	}
	server = client.NewUserServiceClient(load_balance.NewRoundLoadBalance("tenured_store", api.StoreUser, reg))
	if err = server.Start(); err != nil {
		return
	}
	return
}

func TestUserAdd(t *testing.T) {
	server := GetUserService()
	user := &api.User{}
	user.AccountId = 1
	user.AppId = 1
	user.TenantUserId = "haiker"
	user.CloudId = 1
	user.NickName = "haiker"
	err := server.AddUser(user)
	assert.Nil(t, err)
}

func TestUserGet(t *testing.T) {
	server := GetUserService()
	user, err := server.GetByTenantUserId(1, 1, "haiker")
	assert.Nil(t, err)
	t.Log(user)
}

func TestUserGetCloud(t *testing.T) {
	server := GetUserService()
	user, err := server.GetByCloudId(1, 1, 1)
	assert.Nil(t, err)
	t.Log(user)
}

func TestRequestToken(t *testing.T) {
	server := GetUserService()
	rt := new(api.TokenRequest)
	rt.AppId = 1
	rt.AccountId = 1
	rt.CloudId = 1
	rt.IPAddress = "192.168.1.234"

	rp, err := server.RequestLoginToken(rt)
	assert.Nil(t, err)
	t.Log(rp)
}
