package consul

import (
	"github.com/ihaiker/tenured-go-server/commons/logs"
	"github.com/ihaiker/tenured-go-server/commons/registry"
	"github.com/sirupsen/logrus"
	"sync"
)

type ConsulRegistryPlugins struct {
	lock     *sync.Mutex
	registry registry.ServiceRegistry
}

func (this *ConsulRegistryPlugins) Instance(config map[string]string) (*registry.ServerInstance, error) {
	sInstance := &registry.ServerInstance{}

	consulAttr := newInstance()
	consulAttr.Config(config)
	sInstance.PluginAttrs = consulAttr

	return sInstance, nil
}

func (this *ConsulRegistryPlugins) Registry(config registry.PluginConfig) (registry.ServiceRegistry, error) {
	if this.registry != nil {
		return this.registry, nil
	}
	this.lock.Lock()
	defer this.lock.Unlock()

	if this.registry != nil {
		return this.registry, nil
	}

	reg, err := newRegistry(config)
	if err != nil {
		return nil, err
	}
	this.registry = reg
	return reg, nil
}

var logger *logrus.Logger

func init() {
	logger = logs.GetLogger("consul.registry")
	registry.AddPlugins("consul", &ConsulRegistryPlugins{lock: &sync.Mutex{}})
}
