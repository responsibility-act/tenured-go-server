package registry

import "github.com/ihaiker/tenured-go-server/commons/atomic"

type LoadBalance interface {
	Select(obj interface{}) ([]ServerInstance, error)
}

type rangeLoadBalance struct {
	serverName string
	rangeIndex *atomic.AtomicUInt32
	reg        ServiceRegistry
}

func (this *rangeLoadBalance) Select(obj interface{}) ([]ServerInstance, error) {
	currentRangeIndex := this.rangeIndex.GetAndIncrement()
	if ss, err := this.reg.Lookup(this.serverName, nil); err != nil {
		return nil, err
	} else if len(ss) == 0 {
		return ss, err
	} else {
		idx := int(currentRangeIndex % uint32(len(ss)))
		return []ServerInstance{ss[idx]}, nil
	}
}

func NewRangeLoadBalance(serverName string, reg ServiceRegistry) LoadBalance {
	return &rangeLoadBalance{
		serverName: serverName, reg: reg,
		rangeIndex: atomic.NewUint32(0),
	}
}
