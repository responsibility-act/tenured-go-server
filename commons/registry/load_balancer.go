package registry

import "github.com/ihaiker/tenured-go-server/commons/atomic"

type LoadBalanceFragment interface {
	Fragment() int
}

type LoadBalance interface {
	Select(serverName string, obj interface{}, reg ServiceRegistry) ([]ServerInstance, error)
}

type rangeLoadBalance struct {
	rangeIndex *atomic.AtomicUInt32
}

func (this *rangeLoadBalance) Select(serverName string, obj interface{}, registry ServiceRegistry) ([]ServerInstance, error) {
	currendRangeIndex := this.rangeIndex.GetAndIncrement()
	if ss, err := registry.Lookup(serverName, nil); err != nil {
		return nil, err
	} else if len(ss) == 0 {
		return ss, err
	} else {
		idx := int(currendRangeIndex % uint32(len(ss)))
		return []ServerInstance{ss[idx]}, nil
	}
}

func NewRangeLoadBalance() LoadBalance {
	return &rangeLoadBalance{rangeIndex: atomic.NewUint32(0)}
}
