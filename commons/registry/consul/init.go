package consul

import (
	"github.com/ihaiker/tenured-go-server/commons/registry"
)

func init() {
	registry.AddRegistry("consul", ConsulRegistry)
}

func ConsulRegistry(config *registry.PluginConfig) (sInstance *registry.ServerInstance, sRegistry registry.ServiceRegistry, err error) {
	sRegistry, err = newRegistry(config)
	if err != nil {
		return
	}
	sInstance = &registry.ServerInstance{}
	sInstance.PluginAttrs = newInstance()
	return
}
