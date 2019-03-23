package registry

//注册中间插件实现
type RegistryPlugins interface {
	Instance(config map[string]string) (*ServerInstance, error)
	Registry(config PluginConfig) (ServiceRegistry, error)
}

var plugins = map[string]RegistryPlugins{}

func AddPlugins(name string, addPlugins RegistryPlugins) {
	plugins[name] = addPlugins
}

func GetPlugins(name string) (RegistryPlugins, bool) {
	plugin, has := plugins[name]
	return plugin, has
}
