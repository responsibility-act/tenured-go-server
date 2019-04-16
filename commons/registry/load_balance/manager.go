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

func NewLoadBalanceManager(def LoadBalance) *LoadBalanceManager {
	lbm := &LoadBalanceManager{
		store: map[uint16]LoadBalance{},
	}
	lbm.def = def
	return lbm
}
