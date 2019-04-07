package store

import (
	"github.com/ihaiker/tenured-go-server/api/invoke"
	"github.com/ihaiker/tenured-go-server/commons"
	"github.com/ihaiker/tenured-go-server/commons/executors"
	"github.com/ihaiker/tenured-go-server/commons/protocol"
	"github.com/ihaiker/tenured-go-server/services/store/dao"
)

type ServicesInvoke struct {
	//需要执行关闭的服务
	serviceManager *commons.ServiceManager

	config *storeConfig

	server *protocol.TenuredServer
}

func NewServicesWapper(config *storeConfig, server *protocol.TenuredServer) *ServicesInvoke {
	return &ServicesInvoke{
		serviceManager: commons.NewServiceManager(),
		config:         config,
		server:         server,
	}
}

func (this *ServicesInvoke) Start() (err error) {

	{
		executorManager := executors.NewExecutorManager(executors.NewFixedExecutorService(100, 10000))
		accountServer := dao.NewAccountServer(this.config.Data)
		if err = invoke.NewAccountServiceInvoke(this.server, accountServer, executorManager); err != nil {
			return
		}
		this.serviceManager.Add(executorManager)
		this.serviceManager.Add(accountServer)
	}

	return this.serviceManager.Start()
}

func (this *ServicesInvoke) Shutdown(interrupt bool) {
	this.serviceManager.Shutdown(interrupt)
}
