package engine

import (
	"errors"
	"fmt"
	"github.com/ihaiker/tenured-go-server/api"
	"github.com/ihaiker/tenured-go-server/commons/logs"
	"github.com/ihaiker/tenured-go-server/commons/registry"
	"github.com/ihaiker/tenured-go-server/commons/registry/load_balance"
	"github.com/ihaiker/tenured-go-server/commons/runtime"
	"path/filepath"
	"plugin"
)

type StoreEngineConfig struct {
	Type       string            `json:"type" yaml:"type"`
	Attributes map[string]string `json:"attributes" yaml:"attributes"`
}

var logger = logs.GetLogger("plugins")

type StorePlugin interface {
	Account() (api.AccountService, error)
	User() (api.UserService, error)
	Search() (api.SearchService, error)
}

type StoreClientPlugin interface {
	LoadBalance() load_balance.LoadBalance
}

func GetStorePlugin(storeServiceName string, storeConfig *StoreEngineConfig, reg registry.ServiceRegistry) (StorePlugin, error) {
	if storeConfig.Type == "leveldb" {
		return newLevelDBStore(storeServiceName, storeConfig, reg)
	} else {
		return loadStorePlugin(storeServiceName, storeConfig, reg)
	}
}

func GetStoreClientPlugin(storeServiceName string, storeConfig *StoreEngineConfig, reg registry.ServiceRegistry) (StoreClientPlugin, error) {
	if storeConfig.Type == "leveldb" {
		return newLevelDBStoreClient(storeServiceName, reg)
	} else {
		return loadStoreClientPlugin(storeServiceName, storeConfig, reg)
	}
}

func loadStorePlugin(storeServiceName string, config *StoreEngineConfig, reg registry.ServiceRegistry) (StorePlugin, error) {
	pluginFile, _ := filepath.Abs(fmt.Sprintf("%s/../plugins/store/%s.%s", runtime.GetBinDir(), config.Type, runtime.GetLibraryExt()))
	logger.Debug("load store plugins: ", config.Type, " ", pluginFile)
	if p, err := plugin.Open(pluginFile); err != nil {
		return nil, err
	} else if fn, err := p.Lookup("NewStorePlugin"); err != nil {
		return nil, err
	} else if newStorePlugins, match := fn.(func(string, *StoreEngineConfig, registry.ServiceRegistry) (StorePlugin, error)); match {
		return newStorePlugins(storeServiceName, config, reg)
	} else {
		return nil, errors.New("can't found registry plugin in: " + pluginFile)
	}
}

func loadStoreClientPlugin(storeServiceName string, config *StoreEngineConfig, reg registry.ServiceRegistry) (StoreClientPlugin, error) {
	pluginFile, _ := filepath.Abs(fmt.Sprintf("%s/../plugins/store/%s_client.%s", runtime.GetBinDir(), config.Type, runtime.GetLibraryExt()))
	logger.Debug("load storeclient plugins: ", config.Type, " ", pluginFile)
	if p, err := plugin.Open(pluginFile); err != nil {
		return nil, err
	} else if fn, err := p.Lookup("NewStoreClientPlugin"); err != nil {
		return nil, err
	} else if newStorePlugins, match := fn.(func(string, *StoreEngineConfig, registry.ServiceRegistry) (StoreClientPlugin, error)); match {
		return newStorePlugins(storeServiceName, config, reg)
	} else {
		return nil, errors.New("can't found registry plugin in: " + pluginFile)
	}
}
