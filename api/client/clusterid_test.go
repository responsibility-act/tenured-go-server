package client

import (
	"github.com/ihaiker/tenured-go-server/api"
	"github.com/ihaiker/tenured-go-server/commons"
	"github.com/ihaiker/tenured-go-server/commons/registry"
	"github.com/ihaiker/tenured-go-server/commons/registry/cache"
	_ "github.com/ihaiker/tenured-go-server/commons/registry/consul"
	"github.com/kataras/iris/core/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func GetClusterService() (api.ClusterIdService, error) {
	if config, err := registry.ParseConfig("consul://127.0.0.1:8500"); err != nil {
		return nil, err
	} else if plugins, has := registry.GetPlugins(config.Plugin); !has {
		return nil, errors.New("no registry")
	} else if reg, err := plugins.Registry(*config); err != nil {
		return nil, err
	} else {
		return NewClusterIdServiceClient("tenured_store", cache.NewCacheRegistry(reg))
	}
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
