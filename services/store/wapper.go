package store

import (
	"github.com/ihaiker/tenured-go-server/api"
	"github.com/ihaiker/tenured-go-server/commons"
	"github.com/ihaiker/tenured-go-server/commons/executors"
	"github.com/ihaiker/tenured-go-server/commons/protocol"
	"github.com/ihaiker/tenured-go-server/services/store/dao"
)

type ServicesWapper struct {
	//需要执行关闭的服务
	serviceManager *commons.ServiceManager

	executorMap map[string]executors.ExecutorService

	config *storeConfig

	server *protocol.TenuredServer
}

func NewServicesWapper(config *storeConfig) *ServicesWapper {
	return &ServicesWapper{
		serviceManager: commons.NewServiceManager(),
		executorMap:    map[string]executors.ExecutorService{},
		config:         config,
	}
}

func (this *ServicesWapper) SetTCPServer(server *protocol.TenuredServer) {
	this.server = server
}

func (this *ServicesWapper) getExecutors(module string, size, buffer int) executors.ExecutorService {
	if executor, has := this.executorMap[module]; has {
		return executor
	} else {
		executor = executors.NewFixedExecutorService(
			this.config.Executors.Get(module+"Size", size),
			this.config.Executors.Get(module+"Buffer", buffer),
		)
		this.executorMap[module] = executor
		return executor
	}
}

func (this *ServicesWapper) Start() (err error) {

	accountServer := dao.NewAccountServer(this.config.Data)
	invoke := protocol.NewInvoke(this.server, accountServer)
	executor := this.getExecutors("account", 10, 1000)
	if err = invoke.Invoke(api.AccountServiceApply, "Apply", executor); err != nil {
		return
	}
	this.serviceManager.Add(accountServer)

	return this.serviceManager.Start()
}

func (this *ServicesWapper) Shutdown(interrupt bool) {
	//close executor
	for _, v := range this.executorMap {
		v.Shutdown(interrupt)
	}
	this.serviceManager.Shutdown(interrupt)
}
