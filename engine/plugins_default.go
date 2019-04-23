package engine

import (
	"errors"
	"github.com/ihaiker/tenured-go-server/registry"
	"github.com/ihaiker/tenured-go-server/registry/load_balance"

	"github.com/ihaiker/tenured-go-server/api"
	"github.com/ihaiker/tenured-go-server/commons"
	"github.com/ihaiker/tenured-go-server/engine/leveldb"
)

type levelDBStorePlugins struct {
	storeServiceName string
	dataPath         string
}

func (this *levelDBStorePlugins) Account() (api.AccountService, error) {
	return leveldb.NewAccountServer(this.storeServiceName, this.dataPath)
}

func (this *levelDBStorePlugins) User() (api.UserService, error) {
	return leveldb.NewUserServer(this.storeServiceName, this.dataPath)
}

func (this *levelDBStorePlugins) Search() (api.SearchService, error) {
	return leveldb.NewSearchServer(this.storeServiceName, this.dataPath)
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
	loadBalance      load_balance.LoadBalance
}

func (this *levelDBStoreClientPlugins) LoadBalance() load_balance.LoadBalance {
	return this.loadBalance
}

func newLevelDBStoreClient(storeServiceName string, reg registry.ServiceRegistry) (StoreClientPlugin, error) {
	return &levelDBStoreClientPlugins{
		storeServiceName: storeServiceName, reg: reg,
		loadBalance: leveldb.NewLoadBalance(storeServiceName, reg),
	}, nil
}
