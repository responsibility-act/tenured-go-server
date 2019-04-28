package load_balance

import (
	"fmt"
	"github.com/ihaiker/tenured-go-server/registry"

	"github.com/kataras/iris/core/errors"
)

type GlobalLoading struct {
	//当前
	CurrentNode int
	NodeSize    int
	Server      *registry.ServerInstance
}

//是否可以下一个节点
func (this *GlobalLoading) NextNode() bool {
	return (this.CurrentNode == 0 && this.NodeSize == 0) || (this.CurrentNode < this.NodeSize)
}

type noneLoadBalance struct {
	serverName string
	serverTag  []string
	reg        registry.ServiceRegistry
}

func (this *noneLoadBalance) Select(requestCode uint16, obj ...interface{}) ([]*registry.ServerInstance, string, error) {
	if len(obj) < 1 {
		return nil, "", errors.New("global loding is must.")
	}

	if gl, match := obj[0].(*GlobalLoading); !match {
		return nil, "", errors.New("global loading is must.")
	} else if ss, err := this.reg.Lookup(this.serverName, this.serverTag); err != nil {
		return nil, "", err
	} else {
		if len(ss) < gl.CurrentNode+1 {
			return nil, "", errors.New(fmt.Sprintf("global loding out of index. %d > %d", gl.CurrentNode+1, len(ss)))
		}
		defer func() { gl.CurrentNode += 1 }()
		gl.NodeSize = len(ss)
		gl.Server = ss[gl.CurrentNode]
		return []*registry.ServerInstance{ss[gl.CurrentNode]}, "", err
	}
}

func (this *noneLoadBalance) Return(requestCode uint16, key string) {

}

func NewNoneLoadBalance(serverName string, serverTag string, reg registry.ServiceRegistry) LoadBalance {
	lb := &noneLoadBalance{
		serverName: serverName, reg: reg,
	}
	if serverTag == "" {
		lb.serverTag = []string{}
	} else {
		lb.serverTag = []string{serverTag}
	}
	return lb
}
