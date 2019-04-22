package engine

import (
	"errors"
	"fmt"
	"github.com/ihaiker/tenured-go-server/registry"

	"github.com/ihaiker/tenured-go-server/commons/logs"
	"github.com/ihaiker/tenured-go-server/commons/runtime"
	"path/filepath"
	"plugin"
)

var logger = logs.GetLogger("plugins")

func GetStorePlugin(storeServiceName string, storeConfig *StoreEngineConfig) (StorePlugin, error) {
	if storeConfig.Type == "leveldb" {
		return newLevelDBStore(storeServiceName, storeConfig)
	} else {
		return loadStorePlugin(storeServiceName, storeConfig)
	}
}

func loadStorePlugin(storeServiceName string, config *StoreEngineConfig) (StorePlugin, error) {
	pluginFile, _ := filepath.Abs(fmt.Sprintf("%s/../plugins/store/%s.%s", runtime.GetBinDir(), config.Type, runtime.GetLibraryExt()))
	logger.Debug("load store plugins: ", config.Type, " ", pluginFile)
	if p, err := plugin.Open(pluginFile); err != nil {
		return nil, err
	} else if fn, err := p.Lookup("NewStorePlugin"); err != nil {
		return nil, err
	} else if newStorePlugins, match := fn.(StorePluginFunc); match {
		return newStorePlugins(storeServiceName, config)
	} else {
		return nil, errors.New("can't found registry plugin in: " + pluginFile)
	}
}

//获取客户端路由插件
func GetStoreClientPlugin(storeServiceName string, storeConfig *StoreEngineConfig, reg registry.ServiceRegistry) (StoreClientPlugin, error) {
	if storeConfig.Type == "leveldb" {
		return newLevelDBStoreClient(storeServiceName, reg)
	} else {
		return loadStoreClientPlugin(storeServiceName, storeConfig, reg)
	}
}

func loadStoreClientPlugin(storeServiceName string, config *StoreEngineConfig, reg registry.ServiceRegistry) (StoreClientPlugin, error) {
	pluginFile, _ := filepath.Abs(fmt.Sprintf("%s/../plugins/store/%s_client.%s", runtime.GetBinDir(), config.Type, runtime.GetLibraryExt()))
	logger.Debug("load storeclient plugins: ", config.Type, " ", pluginFile)
	if p, err := plugin.Open(pluginFile); err != nil {
		return nil, err
	} else if fn, err := p.Lookup("NewStoreClientPlugin"); err != nil {
		return nil, err
	} else if newStorePlugins, match := fn.(StoreClientPluginFunc); match {
		return newStorePlugins(storeServiceName, config, reg)
	} else {
		return nil, errors.New("can't found registry plugin in: " + pluginFile)
	}
}
