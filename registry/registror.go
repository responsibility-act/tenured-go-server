package registry

/*
	注册中心注册器
*/

const StatusOK = "OK"             //注册
const StatusCritical = "CRITICAL" //临时节点
const StatusDown = "DOWN"         //下线节点

type RegistryNotifyListener func(serverInstances []*ServerInstance)

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
