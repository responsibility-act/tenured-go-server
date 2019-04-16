package tests

import (
	"github.com/ihaiker/tenured-go-server/api/client"
	"github.com/ihaiker/tenured-go-server/commons"
	"github.com/ihaiker/tenured-go-server/commons/registry"
	"github.com/ihaiker/tenured-go-server/commons/registry/load_balance"
	"github.com/ihaiker/tenured-go-server/plugins"
	"github.com/stretchr/testify/assert"
	"testing"
)

func GetClusterService() (server *client.ClusterIdServiceClient, err error) {
	var plugin registry.Plugins
	var reg registry.ServiceRegistry
	if plugin, err = plugins.GetRegistryPlugins("consul://127.0.0.1:8500"); err != nil {
		return
	} else {
		if reg, err = plugin.Registry(); err != nil {
			return
		}
	}
	if server, err = client.NewClusterIdServiceClient(load_balance.NewRoundLoadBalance("tenured_store", "cluster", reg)); err != nil {
		return
	}
	if err = server.Start(); err != nil {
		return
	}
	return
}

func BenchmarkSnowflake(b *testing.B) {
	sf, _ := GetClusterService()
	_ = commons.StartIfService(sf)
	for i := 0; i < b.N; i++ {
		_, _ = sf.Get()
	}
}

func TestClusterId(t *testing.T) {
	server, err := GetClusterService()

	err = commons.StartIfService(server)
	assert.Nil(t, err)

	t.Log(server.Get())
}
