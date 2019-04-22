package engine

import (
	"errors"
	"github.com/ihaiker/tenured-go-server/protocol"
	"github.com/ihaiker/tenured-go-server/registry"
	"github.com/ihaiker/tenured-go-server/registry/load_balance"

	"github.com/ihaiker/tenured-go-server/api"
	"github.com/ihaiker/tenured-go-server/commons"
	"github.com/ihaiker/tenured-go-server/commons/executors"
	"github.com/ihaiker/tenured-go-server/engine/leveldb"
)

type levelDBStorePlugins struct {
	storeServiceName string
	dataPath         string

	reg     registry.ServiceRegistry
	server  *protocol.TenuredServer
	manager executors.ExecutorManager
}

func (this *levelDBStorePlugins) SetRegistry(serviceRegistry registry.ServiceRegistry) {
	this.reg = serviceRegistry
}
func (this *levelDBStorePlugins) SetTenuredServer(server *protocol.TenuredServer) {
	this.server = server
}
func (this *levelDBStorePlugins) SetManager(manager executors.ExecutorManager) {
	this.manager = manager
}

func (this *levelDBStorePlugins) Account() (api.AccountService, error) {
	return leveldb.NewAccountServer(this.dataPath)
}

func (this *levelDBStorePlugins) User() (api.UserService, error) {
	return leveldb.NewUserServer(this.dataPath, this.storeServiceName, this.reg, this.server, this.manager)
}

func (this *levelDBStorePlugins) Search() (api.SearchService, error) {
	return leveldb.NewSearchServer(this.dataPath)
}

func newLevelDBStore(storeServiceName string, config *StoreEngineConfig) (StorePlugin, error) {
	dataPath := commons.NewFile(config.Attributes["dataPath"])
	if !dataPath.Exist() || !dataPath.IsDir() {
		return nil, errors.New("the datapath not found !")
	}

	store := &levelDBStorePlugins{
		storeServiceName: storeServiceName,
		dataPath:         dataPath.GetPath(),
	}
	return store, nil
}

type levelDBStoreClientPlugins struct {
	storeServiceName string
	reg              registry.ServiceRegistry
}

func (this *levelDBStoreClientPlugins) LoadBalance() load_balance.LoadBalance {
	return leveldb.NewLoadBalance(this.storeServiceName, this.reg)
}

func newLevelDBStoreClient(storeServiceName string, reg registry.ServiceRegistry) (StoreClientPlugin, error) {
	return &levelDBStoreClientPlugins{
		storeServiceName: storeServiceName, reg: reg,
	}, nil
}
