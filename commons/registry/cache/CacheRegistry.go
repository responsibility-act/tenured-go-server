package cache

import "github.com/ihaiker/tenured-go-server/commons/registry"

type CacheServiceRegistry struct {
	reg registry.ServiceRegistry

	serverCache map[string][]*registry.ServerInstance
}

func (this *CacheServiceRegistry) Register(serverInstance *registry.ServerInstance) error {
	return this.reg.Register(serverInstance)
}

func (this *CacheServiceRegistry) Unregister(serverId string) error {
	return this.reg.Unregister(serverId)
}

func (this *CacheServiceRegistry) Subscribe(serverName string, listener registry.RegistryNotifyListener) error {
	return this.reg.Subscribe(serverName, func(status registry.RegistionStatus, serverInstances []*registry.ServerInstance) {
		listener(status, serverInstances)
	L1:
		for _, serverInstance := range serverInstances {
			if cacheServerInstances, has := this.serverCache[serverInstance.Name]; has {
				for _, cacheServerInstance := range cacheServerInstances {
					if cacheServerInstance.Id == serverInstance.Id {
						cacheServerInstance.Status = serverInstance.Status
						continue L1
					}
				}
			}
		}
	})
}

func (this *CacheServiceRegistry) Unsubscribe(serverName string, listener registry.RegistryNotifyListener) error {
	return this.reg.Unsubscribe(serverName, listener)
}

func (this *CacheServiceRegistry) Lookup(serverName string, tags []string) ([]*registry.ServerInstance, error) {
	if ss, has := this.serverCache[serverName]; has {
		return ss, nil
	} else {
		ss, err := this.reg.Lookup(serverName, tags)
		if err == nil {
			this.serverCache[serverName] = ss
		}
		return ss, err
	}
}

func NewCacheRegistry(reg registry.ServiceRegistry) registry.ServiceRegistry {
	return &CacheServiceRegistry{
		reg:         reg,
		serverCache: map[string][]*registry.ServerInstance{},
	}
}
