package load_balance

import (
	"github.com/ihaiker/tenured-go-server/commons/atomic"
	"github.com/ihaiker/tenured-go-server/commons/registry"
)

type roundLoadBalance struct {
	serverName string
	serverTag  string
	rangeIndex *atomic.AtomicUInt32
	reg        registry.ServiceRegistry
}

func (this *roundLoadBalance) Select(requestCode uint16, obj ...interface{}) ([]*registry.ServerInstance, string, error) {
	currentRangeIndex := this.rangeIndex.GetAndIncrement()
	if ss, err := this.reg.Lookup(this.serverName, []string{this.serverTag}); err != nil {
		return nil, "", err
	} else if len(ss) == 0 {
		return ss, "", err
	} else {
		idx := int(currentRangeIndex % uint32(len(ss)))
		return []*registry.ServerInstance{ss[idx]}, "", nil
	}
}

func (this *roundLoadBalance) Return(requestCode uint16, key string) {

}

func NewRoundLoadBalance(serverName string, serverTag string, reg registry.ServiceRegistry) LoadBalance {
	return &roundLoadBalance{
		serverName: serverName, serverTag: serverTag, reg: reg,
		rangeIndex: atomic.NewUint32(0),
	}
}
