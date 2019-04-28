package load_balance

import (
	"github.com/ihaiker/tenured-go-server/commons/atomic"
	"github.com/ihaiker/tenured-go-server/protocol"
	"github.com/ihaiker/tenured-go-server/registry"
)

type roundLoadBalance struct {
	serverName string
	serverTag  string
	rangeIndex *atomic.AtomicUInt32
	reg        registry.ServiceRegistry
}

func (this *roundLoadBalance) Select(requestCode uint16, obj ...interface{}) ([]*registry.ServerInstance, string, error) {
	if ss, err := this.reg.Lookup(this.serverName, []string{this.serverTag}); err != nil {
		return nil, "", err
	} else if len(ss) == 0 {
		return ss, "", err
	} else {
		for i := 0; i < len(ss); i++ {
			currentRangeIndex := this.rangeIndex.GetAndIncrement()
			idx := int(currentRangeIndex % uint32(len(ss)))
			if registry.IsOK(ss[idx]) {
				return []*registry.ServerInstance{ss[idx]}, "", nil
			}
		}
		return nil, "", protocol.ErrorRouter()
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
