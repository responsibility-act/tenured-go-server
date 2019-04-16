package load_balance

import "github.com/ihaiker/tenured-go-server/commons/registry"

type LoadBalance interface {
	//注册服务的列表，
	Select(requestCode uint16, obj ...interface{}) (serverInstances []*registry.ServerInstance, regKey string, err error)
	//返回
	Return(requestCode uint16, regKey string)
}
