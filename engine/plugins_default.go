package engine

import (
	"errors"
	"github.com/ihaiker/tenured-go-server/api"
	"github.com/ihaiker/tenured-go-server/commons"
	"github.com/ihaiker/tenured-go-server/commons/registry"
	"github.com/ihaiker/tenured-go-server/commons/registry/load_balance"
	"github.com/ihaiker/tenured-go-server/engine/leveldb"
)

type levelDBStorePlugins struct {
	storeServiceName string
	dataPath         string
	reg              registry.ServiceRegistry
}

func (this *levelDBStorePlugins) Account() (api.AccountService, error) {
	return leveldb.NewAccountServer(this.dataPath)
}

func (this *levelDBStorePlugins) User() (api.UserService, error) {
	return nil, errors.New("not support")
}

func (this *levelDBStorePlugins) Search() (api.SearchService, error) {
	return leveldb.NewSearchServer(this.dataPath)
}

func (this *levelDBStorePlugins) LoadBalance() load_balance.LoadBalance {
	return leveldb.NewLoadBalance()
}

func newLevelDBStore(storeServiceName string, config *StoreEngineConfig, reg registry.ServiceRegistry) (StorePlugins, error) {
	dataPath := commons.NewFile(config.Attributes["dataPath"])
	if !dataPath.Exist() || !dataPath.IsDir() {
		return nil, errors.New("the datapath not found !")
	}

	store := &levelDBStorePlugins{
		storeServiceName: storeServiceName,
		dataPath:         dataPath.GetPath(),
		reg:              reg,
	}
	return store, nil
}
