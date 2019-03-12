package registry

import (
	"errors"
)

//注册中间插件实现
type RegistryPlugins func(*PluginConfig) (*ServerInstance, ServiceRegistry, error)

var plugins = map[string]RegistryPlugins{}

func AddRegistry(name string, reg RegistryPlugins) {
	plugins[name] = reg
}

func GetRegistry(config *PluginConfig) (*ServerInstance, ServiceRegistry, error) {
	name := config.Plugin
	if plugin, has := plugins[name]; !has {
		return nil, nil, errors.New("Can't found the registry named " + name)
	} else {
		return plugin(config)
	}
}
