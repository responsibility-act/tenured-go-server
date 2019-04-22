package plugins

import (
	"errors"
	"fmt"
	"github.com/ihaiker/tenured-go-server/commons/logs"
	"github.com/ihaiker/tenured-go-server/commons/runtime"
	"github.com/ihaiker/tenured-go-server/registry"
	"github.com/ihaiker/tenured-go-server/registry/consul"
	"path/filepath"
	"plugin"
)

var logger = logs.GetLogger("plugins")

func GetRegistryPlugins(registryConfig string) (registry.Plugins, error) {
	if config, err := registry.ParseConfig(registryConfig); err != nil {
		return nil, err
	} else if config.Plugin == "consul" {
		return consul.NewRegistryPlugins(config)
	} else {
		return loadPluginRegistry(config)
	}
}

func loadPluginRegistry(config *registry.PluginConfig) (registry.Plugins, error) {
	pluginFile, _ := filepath.Abs(fmt.Sprintf("%s/../plugins/registry/%s.%s", runtime.GetBinDir(), config.Plugin, runtime.GetLibraryExt()))
	logger.Debug("load registry: ", config.Plugin, " ", pluginFile)
	if p, err := plugin.Open(pluginFile); err != nil {
		return nil, err
	} else if fn, err := p.Lookup("NewRegistryPlugins"); err != nil {
		return nil, err
	} else if newFn, match := fn.(func(*registry.PluginConfig) (registry.Plugins, error)); match {
		return newFn(config)
	} else {
		return nil, errors.New("can't found registry plugin in: " + pluginFile)
	}
}
