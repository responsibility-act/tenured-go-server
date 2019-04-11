package registry

/*
	注册中心注册器
*/

type RegistionStatus int

func (this RegistionStatus) String() string {
	switch this {
	case REGISTER:
		return "register"
	case UNREGISTER:
		return "unregister"
	}
	return "unknown"
}

const REGISTER = RegistionStatus(0)
const UNREGISTER = RegistionStatus(1)

type RegistryNotifyListener func(status RegistionStatus, serverInstances []*ServerInstance)

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
