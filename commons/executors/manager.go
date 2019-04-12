package executors

import "github.com/ihaiker/tenured-go-server/commons"

type ExecutorManager interface {
	commons.Service

	Get(name string) ExecutorService
	Fix(name string, size, buffer int) ExecutorService
	Single(name string, buffer int) ExecutorService
}

type defExecutorManager struct {
	def         ExecutorService
	executorMap map[string]ExecutorService
}

func (this *defExecutorManager) Get(module string) ExecutorService {
	if executor, has := this.executorMap[module]; has {
		return executor
	} else {
		return this.def
	}
}

func (this *defExecutorManager) Fix(module string, size, buffer int) ExecutorService {
	if executor, has := this.executorMap[module]; has {
		return executor
	} else {
		executor = NewFixedExecutorService(size, buffer)
		this.executorMap[module] = executor
		return executor
	}
}

func (this *defExecutorManager) Single(module string, buffer int) ExecutorService {
	if executor, has := this.executorMap[module]; has {
		return executor
	} else {
		executor = NewSingleExecutorService(buffer)
		this.executorMap[module] = executor
		return executor
	}
}

func (this *defExecutorManager) Start() error {
	return nil
}

func (this *defExecutorManager) Shutdown(interrupt bool) {
	for _, v := range this.executorMap {
		v.Shutdown(interrupt)
	}
}

func NewExecutorManager(def ExecutorService) ExecutorManager {
	return &defExecutorManager{
		def:         def,
		executorMap: map[string]ExecutorService{},
	}
}
