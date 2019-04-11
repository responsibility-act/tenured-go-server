package registry

import "github.com/ihaiker/tenured-go-server/commons/atomic"

type roundLoadBalance struct {
	serverName string
	rangeIndex *atomic.AtomicUInt32
	reg        ServiceRegistry
}

func (this *roundLoadBalance) Select(obj ...interface{}) ([]*ServerInstance, string, error) {
	currentRangeIndex := this.rangeIndex.GetAndIncrement()
	if ss, err := this.reg.Lookup(this.serverName, nil); err != nil {
		return nil, "", err
	} else if len(ss) == 0 {
		return ss, "", err
	} else {
		idx := int(currentRangeIndex % uint32(len(ss)))
		return []*ServerInstance{ss[idx]}, "", nil
	}
}

func (this *roundLoadBalance) Return(key string) {

}

func NewRoundLoadBalance(serverName string, reg ServiceRegistry) LoadBalance {
	return &roundLoadBalance{
		serverName: serverName, reg: reg,
		rangeIndex: atomic.NewUint32(0),
	}
}
