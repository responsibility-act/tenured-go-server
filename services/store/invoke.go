package store

import (
	"github.com/ihaiker/tenured-go-server/api/invoke"
	"github.com/ihaiker/tenured-go-server/commons"
	"github.com/ihaiker/tenured-go-server/commons/executors"
	"github.com/ihaiker/tenured-go-server/commons/protocol"
	"github.com/ihaiker/tenured-go-server/services/store/dao"
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

func (this *ServicesInvokeManager) Start() (err error) {
	{
		accountServer := dao.NewAccountServer(this.config.Data)
		if err = invoke.NewAccountServiceInvoke(this.server, accountServer, this.executorManager); err != nil {
			return
		}
		this.serviceManager.Add(this.executorManager, accountServer)
	}

	return this.serviceManager.Start()
}

func (this *ServicesInvokeManager) Shutdown(interrupt bool) {
	this.serviceManager.Shutdown(interrupt)
}
