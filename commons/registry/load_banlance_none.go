package registry

import (
	"fmt"
	"github.com/kataras/iris/core/errors"
)

type GlobalLoading struct {
	//当前
	CurrentNode int
	NodeSize    int
	Server      *ServerInstance
}

//是否可以下一个节点
func (this *GlobalLoading) NextNode() bool {
	return (this.CurrentNode == 0 && this.NodeSize == 0) || (this.CurrentNode < this.NodeSize)
}

type noneLoadBalance struct {
	serverName string
	reg        ServiceRegistry
}

func (this *noneLoadBalance) Select(obj ...interface{}) ([]*ServerInstance, string, error) {
	if len(obj) < 2 {
		return nil, "", errors.New("global loding is must.")
	}

	if gl, match := obj[1].(*GlobalLoading); !match {
		return nil, "", errors.New("global loading is must.")
	} else if ss, err := this.reg.Lookup(this.serverName, nil); err != nil {
		return nil, "", err
	} else {
		if len(ss) < gl.CurrentNode+1 {
			return nil, "", errors.New(fmt.Sprintf("global loding out of index. %d > %d", gl.CurrentNode+1, len(ss)))
		}
		defer func() { gl.CurrentNode += 1 }()
		gl.NodeSize = len(ss)
		gl.Server = ss[gl.CurrentNode]
		return []*ServerInstance{ss[gl.CurrentNode]}, "", err
	}
}

func (this *noneLoadBalance) Return(key string) {

}

func NewNoneLoadBalance(serverName string, reg ServiceRegistry) LoadBalance {
	return &noneLoadBalance{
		serverName: serverName, reg: reg,
	}
}
