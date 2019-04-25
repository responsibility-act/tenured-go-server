package registry

import (
	"fmt"
)

/*
	注册中心注册器
*/

const StatusOK = "OK"             //注册
const StatusCritical = "CRITICAL" //临时节点
const StatusDown = "DOWN"         //下线节点

type RegistryNotifyListener func(serverInstances []*ServerInstance)

//由于监听都是定义的fn，并且需要存储，而slice中不可存储fn，
// 所以就需要存入map中，而放入map里面就需要寻找一个可以充当key的值，
// 当初第一想法是使用uintptr，使用reflect.Value(fn).Pointer()，然而测试并非如此，可以查看DOC，
// 另辟蹊径查到 %p可以打印地址，好了就他了
func NotifyPointer(notifyFn RegistryNotifyListener) string {
	p := &notifyFn
	return fmt.Sprintf("%p", p)
}

type ServiceRegistry interface {
	//想注册中心注册服务
	Register(serverInstance *ServerInstance) error

	//从注册中心删除注册
	Unregister(serverId string) error

	//订阅服务改变
	Subscribe(serverName string, listener RegistryNotifyListener) error

	//取消服务订阅
	Unsubscribe(serverName string, listener RegistryNotifyListener) error

	//发现服务内容
	Lookup(serverName string, tags []string) ([]*ServerInstance, error)
}
