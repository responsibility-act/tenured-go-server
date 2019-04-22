package consul

import (
	"github.com/ihaiker/tenured-go-server/commons/logs"
	"github.com/ihaiker/tenured-go-server/registry"
	"github.com/sirupsen/logrus"
	"sync"
)

type ConsulRegistryPlugins struct {
	lock     *sync.Mutex
	registry registry.ServiceRegistry
	config   *registry.PluginConfig
}

func (this *ConsulRegistryPlugins) Instance(config map[string]string) (*registry.ServerInstance, error) {
	sInstance := &registry.ServerInstance{}

	consulAttr := newInstance()
	consulAttr.Config(config)
	sInstance.PluginAttrs = consulAttr

	return sInstance, nil
}

func (this *ConsulRegistryPlugins) Registry() (registry.ServiceRegistry, error) {
	if this.registry != nil {
		return this.registry, nil
	}
	this.lock.Lock()
	defer this.lock.Unlock()

	if this.registry != nil {
		return this.registry, nil
	}

	if reg, err := newRegistry(this.config); err != nil {
		return nil, err
	} else {
		this.registry = reg
		return reg, nil
	}
}

var logger *logrus.Logger

func init() {
	logger = logs.GetLogger("consul")
}

func NewRegistryPlugins(config *registry.PluginConfig) (registry.Plugins, error) {
	return &ConsulRegistryPlugins{
		lock:   new(sync.Mutex),
		config: config,
	}, nil
}
