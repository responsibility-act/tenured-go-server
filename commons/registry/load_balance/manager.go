package load_balance

import (
	"github.com/ihaiker/tenured-go-server/commons"
	"github.com/ihaiker/tenured-go-server/commons/registry"
)

type LoadBalanceManager struct {
	def   LoadBalance
	store map[uint16]LoadBalance
}

func (this *LoadBalanceManager) AddLoadBalance(requestCode uint16, lb LoadBalance) {
	this.store[requestCode] = lb
}

func (this *LoadBalanceManager) Select(requestCode uint16, obj ...interface{}) (serverInstances []*registry.ServerInstance, regKey string, err error) {
	if lb, has := this.store[requestCode]; has {
		serverInstances, regKey, err = lb.Select(requestCode, obj)
	} else {
		serverInstances, regKey, err = this.def.Select(requestCode, obj)
	}
	if commons.IsNil(err) {
		return
	}
	return
}

func (this *LoadBalanceManager) Return(requestCode uint16, regKey string) {
	if lb, has := this.store[requestCode]; has {
		lb.Return(requestCode, regKey)
	} else {
		this.def.Return(requestCode, regKey)
	}
}

func (this *LoadBalanceManager) Start() error {
	sets := map[LoadBalance]int{} //因为一个lb会多次注册，记录已经启动的，不然会多次调用启动
	for _, v := range this.store {
		if _, has := sets[v]; has {
			continue
		}
		sets[v] = 1
		if err := commons.StartIfService(v); err != nil {
			return err
		}
	}
	return nil
}

func (this *LoadBalanceManager) Shutdown(interrupt bool) {
	sets := map[LoadBalance]int{} //同Start处理逻辑一致
	for _, v := range this.store {
		if _, has := sets[v]; has {
			continue
		}
		sets[v] = 1
		commons.ShutdownIfService(v, interrupt)
	}
}

func NewLoadBalanceManager(def LoadBalance) *LoadBalanceManager {
	lbm := &LoadBalanceManager{
		store: map[uint16]LoadBalance{},
	}
	lbm.def = def
	return lbm
}
