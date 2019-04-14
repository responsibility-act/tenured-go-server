package store

import (
	"github.com/ihaiker/tenured-go-server/api/invoke"
	"github.com/ihaiker/tenured-go-server/commons"
	"github.com/ihaiker/tenured-go-server/commons/executors"
	"github.com/ihaiker/tenured-go-server/commons/protocol"
	"github.com/ihaiker/tenured-go-server/engine"
)

type ServicesInvokeManager struct {
	//需要执行关闭的服务
	serviceManager *commons.ServiceManager

	config *storeConfig

	server *protocol.TenuredServer

	executorManager executors.ExecutorManager
}

func NewServicesInvokeManager(config *storeConfig, server *protocol.TenuredServer, executorManager executors.ExecutorManager) *ServicesInvokeManager {
	return &ServicesInvokeManager{
		serviceManager:  commons.NewServiceManager(),
		config:          config,
		server:          server,
		executorManager: executorManager,
	}
}

func (this *ServicesInvokeManager) Start() error {
	storePlugins, err := engine.GetStorePlugins(this.config.Store)
	if err != nil {
		return err
	}

	{
		accountServer := storePlugins.Account()
		if err := invoke.NewAccountServiceInvoke(this.server, accountServer, this.executorManager); err != nil {
			return err
		}
		this.serviceManager.Add(this.executorManager, accountServer)
	}

	return this.serviceManager.Start()
}

func (this *ServicesInvokeManager) Shutdown(interrupt bool) {
	this.serviceManager.Shutdown(interrupt)
}
