package store

import (
	"github.com/ihaiker/tenured-go-server/api"
	"github.com/ihaiker/tenured-go-server/api/invoke"
	"github.com/ihaiker/tenured-go-server/commons"
	"github.com/ihaiker/tenured-go-server/commons/executors"
	"github.com/ihaiker/tenured-go-server/commons/protocol"
	"github.com/ihaiker/tenured-go-server/commons/registry"
	"github.com/ihaiker/tenured-go-server/engine"
)

type ServicesInvokeManager struct {
	//需要执行关闭的服务
	config *storeConfig

	reg registry.ServiceRegistry

	server *protocol.TenuredServer

	executorManager executors.ExecutorManager

	storePlugins engine.StorePlugin

	serverManager *commons.ServiceManager
}

func NewServicesInvokeManager(config *storeConfig, reg registry.ServiceRegistry, server *protocol.TenuredServer, executorManager executors.ExecutorManager) *ServicesInvokeManager {
	return &ServicesInvokeManager{
		reg:             reg,
		config:          config,
		server:          server,
		executorManager: executorManager,
		serverManager:   commons.NewServiceManager(),
	}
}

func (this *ServicesInvokeManager) Start() (err error) {
	storeServerName := this.config.Prefix + "_store"
	this.storePlugins, err = engine.GetStorePlugin(storeServerName, this.config.Engine, this.reg)
	if err != nil {
		return err
	}

	if this.config.HasStore(api.StoreAccount) {
		if service, err := this.storePlugins.Account(); err != nil {
			return err
		} else if err := invoke.NewAccountServiceInvoke(this.server, service, this.executorManager); err != nil {
			return err
		} else {
			this.serverManager.Add(service)
		}
	}

	if this.config.HasStore(api.StoreSearch) {
		if service, err := this.storePlugins.Search(); err != nil {
			return err
		} else if err := invoke.NewSearchServiceInvoke(this.server, service, this.executorManager); err != nil {
			return err
		} else {
			this.serverManager.Add(service)
		}
	}

	return this.serverManager.Start()
}

func (this *ServicesInvokeManager) Shutdown(interrupt bool) {
	this.serverManager.Shutdown(interrupt)
}
