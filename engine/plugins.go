package engine

import (
	"errors"
	"github.com/ihaiker/tenured-go-server/api"
	"github.com/ihaiker/tenured-go-server/engine/leveldb"
)

type StoreConfig struct {
	Type       string            `json:"type" yaml:"type"`
	Attributes map[string]string `json:"attributes" yaml:"attributes"`
}

type StorePlugins interface {
	Account() api.AccountService
	User() api.UserService
}

type levelDBStorePlugins struct {
	account api.AccountService
	user    api.UserService
}

func (this *levelDBStorePlugins) Account() api.AccountService {
	return this.account
}

func (this *levelDBStorePlugins) User() api.UserService {
	return this.user
}

func newLevelDBStore(config *StoreConfig) (StorePlugins, error) {
	store := &levelDBStorePlugins{}
	store.account = leveldb.NewAccountServer(config.Attributes["dataPath"])
	return store, nil
}

func GetStorePlugins(storeConfig *StoreConfig) (StorePlugins, error) {
	if storeConfig.Type == "leveldb" {
		return newLevelDBStore(storeConfig)
	} else {
		return nil, errors.New("not support store: " + storeConfig.Type)
	}
}
