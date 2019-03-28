package consul

import (
	"fmt"
	"github.com/emirpasic/gods/sets/hashset"
	"github.com/hashicorp/consul/api"
	"github.com/ihaiker/tenured-go-server/commons/registry"
	"github.com/sirupsen/logrus"
	"net"
	"strconv"
)

type subscribeInfo struct {
	listeners *hashset.Set
	services  map[string]registry.ServerInstance
	closeChan chan struct{}
}

func (this *subscribeInfo) close() {
	close(this.closeChan)
}

type ConsulServiceRegistry struct {
	client *api.Client
	config *ConsulConfig

	subscribes map[string]*subscribeInfo
}

func (this *ConsulServiceRegistry) Start() error {
	return nil
}

func (this *ConsulServiceRegistry) Shutdown(interrupt bool) {
	for name, ch := range this.subscribes {
		ch.close()
		delete(this.subscribes, name)
	}
}

func (this *ConsulServiceRegistry) Register(serverInstance registry.ServerInstance) error {
	logrus.Infof("To register %s(%s) : %s", serverInstance.Name, serverInstance.Address, serverInstance.Id)
	attrs := serverInstance.PluginAttrs.(*ConsulServerAttrs)
	if host, portStr, err := net.SplitHostPort(serverInstance.Address); err != nil {
		return err
	} else if port, err := strconv.Atoi(portStr); err != nil {
		return err
	} else {
		check := &api.AgentServiceCheck{ // 健康检查
			Interval:                       attrs.Interval,
			Timeout:                        attrs.RequestTimeout,
			DeregisterCriticalServiceAfter: attrs.Deregister,
		}
		switch attrs.CheckType {
		case "http":
			check.HTTP = "http://" + serverInstance.Address + "/" + attrs.Health
		case "tcp":
			check.TCP = serverInstance.Address
		}

		reg := &api.AgentServiceRegistration{
			ID:      serverInstance.Id,   // 服务节点的名称
			Name:    serverInstance.Name, // 服务名称
			Meta:    serverInstance.Metadata,
			Address: host, Port: port, // 服务 IP:端口
			Check: check,
		}
		return this.client.Agent().ServiceRegister(reg)
	}
}

func (this *ConsulServiceRegistry) Unregister(serverId string) error {
	logrus.Info("To Unregister ", serverId)
	return this.client.Agent().ServiceDeregister(serverId)
}

func (this *ConsulServiceRegistry) convertService(serverName string, service *api.ServiceEntry) registry.ServerInstance {
	status := service.Checks.AggregatedStatus()
	if status == api.HealthPassing {
		status = "OK"
	}
	return registry.ServerInstance{
		Id:       service.Service.ID,
		Name:     serverName,
		Metadata: service.Service.Meta,
		Address:  fmt.Sprintf("%s:%d", service.Service.Address, service.Service.Port),
		Tags:     service.Service.Tags,
		Status:   status,
	}
}

func (this *ConsulServiceRegistry) loadSubscribeHealth(serverName string) {
	defer func() {
		if e := recover(); e != nil {
			logrus.Warnf("close subscribe(%s) error: %v", serverName, e)
		}
	}()
	logrus.Debug("start loop load subscribe server health:", serverName)

	waitIndex := uint64(0)
	healthWaitTime := this.config.HealthWaitTime()
	failWaitTime := this.config.HealthFailWaitTime()
	waitTime := healthWaitTime

	register := make([]registry.ServerInstance, 0)
	deregister := make([]registry.ServerInstance, 0)

	for {
		subInfo, has := this.subscribes[serverName]
		if !has {
			return
		}

		select {
		case <-subInfo.closeChan:
			return
		default:
			services, metainfo, err := this.client.Health().Service(serverName, "", false,
				&api.QueryOptions{
					WaitIndex: waitIndex, //同步点，这个调用将一直阻塞，直到有新的更新,
					WaitTime:  waitTime,  //此次请求等待时间，此处设置防止携程阻死
					//UseCache:  true, MaxAge:time.Second*5
				})
			if err != nil || waitIndex == metainfo.LastIndex {
				waitTime = healthWaitTime
				continue
			}

			subInfo, has = this.subscribes[serverName]
			if !has {
				return
			}

			register = register[:0]
			deregister = deregister[:0]
			if subInfo.services == nil {
				subInfo.services = map[string]registry.ServerInstance{}
				for _, s := range services {
					subInfo.services[s.Service.ID] = this.convertService(serverName, s)
				}
			} else {
				currentServices := map[string]registry.ServerInstance{}

				for _, s := range services {
					current := this.convertService(serverName, s)
					if current.Status != api.HealthPassing {
						waitTime = failWaitTime
					}
					if old, has := subInfo.services[s.Service.ID]; !has || current.Status != old.Status {
						register = append(register, current)
					}
					currentServices[s.Service.ID] = current
					delete(subInfo.services, s.Service.ID)
				}

				for _, s := range subInfo.services {
					s.Status = "deregister"
					deregister = append(deregister, s)
				}

				for _, v := range subInfo.listeners.Values() {
					if len(register) > 0 {
						v.(registry.RegistryNotifyListener).
							OnNotify(registry.REGISTER, register)
					}
					if len(deregister) > 0 {
						v.(registry.RegistryNotifyListener).
							OnNotify(registry.UNREGISTER, deregister)
					}
				}
				subInfo.services = currentServices
			}
			waitIndex = metainfo.LastIndex
		}
	}
}

func (this *ConsulServiceRegistry) Subscribe(serverName string, listener registry.RegistryNotifyListener) error {
	if this.addSubscribe(serverName, listener) {
		go this.loadSubscribeHealth(serverName)
	}
	return nil
}

func (this *ConsulServiceRegistry) Unsubscribe(serverName string, listener registry.RegistryNotifyListener) error {
	if this.removeSubscribe(serverName, listener) {
		if sub, has := this.subscribes[serverName]; has {
			sub.close()
			delete(this.subscribes, serverName)
		}
	}
	return nil
}

func (this *ConsulServiceRegistry) Lookup(serverName string, tags []string) ([]registry.ServerInstance, error) {
	if services, _, err := this.client.Health().
		ServiceMultipleTags(serverName, tags, false, &api.QueryOptions{}); err != nil {
		return nil, err
	} else {
		serverInstances := make([]registry.ServerInstance, len(services))
		for i := 0; i < len(services); i++ {
			serverInstances[i] = this.convertService(serverName, services[i])
		}
		return serverInstances, nil
	}
}

func (this *ConsulServiceRegistry) getOrCreateSubscribe(name string) *subscribeInfo {
	if subInfo, has := this.subscribes[name]; !has {
		subInfo = &subscribeInfo{
			listeners: hashset.New(),
			services:  nil,
			closeChan: make(chan struct{}),
		}
		this.subscribes[name] = subInfo
	}
	return this.subscribes[name]
}

//@return 返回是否是此服务的第一个监听器
func (this *ConsulServiceRegistry) addSubscribe(name string, listener registry.RegistryNotifyListener) bool {
	sets := this.getOrCreateSubscribe(name)
	sets.listeners.Add(listener)
	return sets.listeners.Size() == 1
}

//@return 是否是次服务的最后一个监听器
func (this *ConsulServiceRegistry) removeSubscribe(name string, listener registry.RegistryNotifyListener) bool {
	sets := this.getOrCreateSubscribe(name)
	sets.listeners.Remove(listener)
	return sets.listeners.Size() == 0
}

func newRegistry(pluginConfig registry.PluginConfig) (*ConsulServiceRegistry, error) {
	config := &ConsulConfig{config: pluginConfig}
	serviceRegistry := &ConsulServiceRegistry{
		config:     config,
		subscribes: map[string]*subscribeInfo{},
	}

	consulApiCfg := api.DefaultConfig()
	consulApiCfg.Scheme = config.Scheme()
	consulApiCfg.Address = config.Address()
	consulApiCfg.Datacenter = config.Datacenter()
	consulApiCfg.Token = config.Token()

	if client, err := api.NewClient(consulApiCfg); err != nil {
		return nil, err
	} else {
		serviceRegistry.client = client
	}
	return serviceRegistry, nil
}
