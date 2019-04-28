package tests

import (
	"github.com/ihaiker/tenured-go-server/api"
	"github.com/ihaiker/tenured-go-server/api/client"
	"github.com/ihaiker/tenured-go-server/registry"
	"github.com/ihaiker/tenured-go-server/registry/load_balance"
	"github.com/ihaiker/tenured-go-server/registry/plugins"
)

func GetLinkerService() (server *client.UserServiceClient) {
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

	server = client.NewUserServiceClient(load_balance.NewNoneLoadBalance("tenured_store", api.StoreLinker, reg))
	if err = server.Start(); err != nil {
		return
	}
	return
}
