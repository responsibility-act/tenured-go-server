package engine

import (
	"github.com/ihaiker/tenured-go-server/api"
	"github.com/ihaiker/tenured-go-server/commons/executors"
	"github.com/ihaiker/tenured-go-server/commons/protocol"
	"github.com/ihaiker/tenured-go-server/commons/registry"
	"github.com/ihaiker/tenured-go-server/commons/registry/load_balance"
)

type StoreEngineConfig struct {
	Type       string            `json:"type" yaml:"type"`
	Attributes map[string]string `json:"attributes" yaml:"attributes"`
}

//存储插件
type StorePlugin interface {
	Account() (api.AccountService, error)
	User() (api.UserService, error)
	Search() (api.SearchService, error)
}
type StorePluginFunc func(storeServiceName string, config *StoreEngineConfig) (StorePlugin, error)

//客户端路由插件
type StoreClientPlugin interface {
	LoadBalance() load_balance.LoadBalance
}
type StoreClientPluginFunc func(storeServiceName string, config *StoreEngineConfig, reg registry.ServiceRegistry) (StoreClientPlugin, error)

//注册组件Aware
type RegistryAware interface {
	SetRegistry(serviceRegistry registry.ServiceRegistry)
}

//Server组件Aware
type TenuredServerAware interface {
	SetTenuredServer(server *protocol.TenuredServer)
}

//执行组件Aware
type ExecutorManagerAware interface {
	SetManager(manager executors.ExecutorManager)
}
