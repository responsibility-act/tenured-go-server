package consul

import (
	"fmt"
	"github.com/emirpasic/gods/sets/hashset"
	"github.com/hashicorp/consul/api"
	"github.com/ihaiker/tenured-go-server/commons/registry"
	"github.com/sirupsen/logrus"
	"net"
	"strconv"
	"sync"
	"time"
)

type ConsulServiceRegistry struct {
	client *api.Client
	config *registry.PluginConfig

	lock               *sync.Mutex
	subscribes         map[string]*hashset.Set
	subscribeCloseChan map[string]chan struct{}
}

func (this *ConsulServiceRegistry) Start() error {
	return nil
}

func (this *ConsulServiceRegistry) Shutdown() {
	this.lock.Lock()
	defer this.lock.Unlock()
	for name, ch := range this.subscribeCloseChan {
		close(ch)
		delete(this.subscribeCloseChan, name)
	}
}

func (this *ConsulServiceRegistry) Register(serverInstance registry.ServerInstance) error {
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
			check.HTTP = serverInstance.Address
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
	return this.client.Agent().ServiceDeregister(serverId)
}

func (this *ConsulServiceRegistry) loadHealth(serverName string) {
	waitIndex := uint64(0)
	waitTime := time.Second * time.Duration(this.config.GetInt("healthTimeout", 5))
	for {
		select {
		case <-this.subscribeCloseChan[serverName]:
			return
		default:
			services, metainfo, err := this.client.Health().Service(serverName, "", false, &api.QueryOptions{
				WaitIndex: waitIndex, // 同步点，这个调用将一直阻塞，直到有新的更新,
				WaitTime:  waitTime,  //或者有等候时间到
				//UseCache:  true,
			})
			if err != nil || waitIndex == metainfo.LastIndex {
				continue
			}

			for _, s := range services {
				createIndex := s.Service.CreateIndex
				modifyIndex := s.Service.ModifyIndex
				if createIndex > waitIndex || modifyIndex > waitIndex {
					//新服务上线
					logrus.Infof("服务变动: %s %s:%d  %s", s.Service.ID, s.Service.Address, s.Service.Port, s.Checks.AggregatedStatus())
				}
			}

			waitIndex = metainfo.LastIndex
		}
	}
}

func (this *ConsulServiceRegistry) Subscribe(serverName string, listener registry.RegistryNotifyListener) error {
	this.lock.Lock()
	defer this.lock.Unlock()
	if this.addSubscribe(serverName, listener) {
		this.subscribeCloseChan[serverName] = make(chan struct{})
		go this.loadHealth(serverName)
	}
	return nil
}

func (this *ConsulServiceRegistry) Unsubscribe(serverName string, listener registry.RegistryNotifyListener) error {
	this.lock.Lock()
	defer this.lock.Unlock()
	if this.removeSubscribe(serverName, listener) {
		if cChan, has := this.subscribeCloseChan[serverName]; has {
			close(cChan)
			delete(this.subscribeCloseChan, serverName)
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
			service := services[i]

			status := service.Checks.AggregatedStatus()
			if status == api.HealthPassing {
				status = "OK"
			}

			serverInstances[i] = registry.ServerInstance{
				Id:       service.Service.ID,
				Name:     serverName,
				Metadata: service.Service.Meta,
				Address:  fmt.Sprintf("%s:%d", service.Service.Address, service.Service.Port),
				Tags:     service.Service.Tags,
				Status:   status,
			}
		}
		return serverInstances, nil
	}
}

func (this *ConsulServiceRegistry) getSubscribe(name string) *hashset.Set {
	if sets, has := this.subscribes[name]; !has {
		sets = hashset.New()
		this.subscribes[name] = sets
	}
	return this.subscribes[name]
}

//@return 返回是否是此服务的第一个监听器
func (this *ConsulServiceRegistry) addSubscribe(name string, listener registry.RegistryNotifyListener) bool {
	sets := this.getSubscribe(name)
	sets.Add(listener)
	return sets.Size() == 1
}

//@return 是否是次服务的最后一个监听器
func (this *ConsulServiceRegistry) removeSubscribe(name string, listener registry.RegistryNotifyListener) bool {
	sets := this.getSubscribe(name)
	sets.Remove(listener)
	return sets.Size() == 0
}

func newRegistry(pluginConfig *registry.PluginConfig) (*ConsulServiceRegistry, error) {
	serviceRegistry := &ConsulServiceRegistry{
		config:             pluginConfig,
		lock:               &sync.Mutex{},
		subscribes:         map[string]*hashset.Set{},
		subscribeCloseChan: map[string]chan struct{}{},
	}

	consulConfig := api.DefaultConfig()
	consulConfig.Scheme = pluginConfig.Get("scheme", "http")
	consulConfig.Address = pluginConfig.Address[0]
	pluginConfig.Apply("datacenter", func(value string) {
		consulConfig.Datacenter = value
	})
	pluginConfig.Apply("token", func(value string) {
		consulConfig.Token = value
	})

	if client, err := api.NewClient(consulConfig); err != nil {
		return nil, err
	} else {
		serviceRegistry.client = client
	}
	return serviceRegistry, nil
}
