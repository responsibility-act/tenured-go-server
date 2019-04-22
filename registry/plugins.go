package registry

//注册中间插件实现
type Plugins interface {
	Instance(config map[string]string) (*ServerInstance, error)
	Registry() (ServiceRegistry, error)
}
