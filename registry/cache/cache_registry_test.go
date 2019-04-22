package cache

import (
	"github.com/ihaiker/tenured-go-server/registry"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestCacheServiceRegistry(t *testing.T) {
	consulPlugin, err := registry.GetRegistryPlugins("consul://127.0.0.1:8500")
	assert.Nil(t, err)

	reg, err := consulPlugin.Registry()
	assert.Nil(t, err)

	cache := registry.NewCacheRegistry(reg)

	w := sync.WaitGroup{}
	w.Add(1)

	_ = cache.Subscribe("tenured_store", func(serverInstances []*registry.ServerInstance) {
		for k, v := range serverInstances {
			t.Log(k, v)
			if v.Status == registry.StatusOK {
				w.Done()
			}
		}
	})
	w.Wait()

	ss, err := cache.Lookup("tenured_store", nil)
	assert.Nil(t, err)

	for k, v := range ss {
		t.Log(k, v)
	}
}
