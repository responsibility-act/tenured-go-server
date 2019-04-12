package dao

import (
	"encoding/json"
	"github.com/ihaiker/tenured-go-server/api"
	"github.com/ihaiker/tenured-go-server/commons"
	"github.com/ihaiker/tenured-go-server/commons/protocol"
	"github.com/ihaiker/tenured-go-server/commons/registry"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/comparer"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"os"
)

var readOptions = &opt.ReadOptions{}
var writeOptions = &opt.WriteOptions{Sync: true}

const levelDBNotFound = "leveldb: not found"

type AccountServer struct {
	dataPath string
	data     *leveldb.DB
}

func NewAccountServer(dataPath string) *AccountServer {
	return &AccountServer{dataPath: dataPath + "/store/account"}
}

func (this *AccountServer) Apply(account *api.Account) *protocol.TenuredError {
	logger.Infof("申请用户：%v", account)
	if storeAccount, err := this.Get(account.Id); err != nil {
		return err
	} else if storeAccount != nil {
		return api.ErrAccountExists
	}
	account.Status = api.AccountStatusApply

	if bs, err := json.Marshal(account); err != nil {
		return protocol.ErrorDB(err)
	} else if err := this.data.Put(commons.UInt64(account.Id), bs, writeOptions); err != nil {
		return protocol.ErrorDB(err)
	}

	return nil
}

func (this *AccountServer) Get(id uint64) (*api.Account, *protocol.TenuredError) {
	logger.Debug(" get user: ", id)

	if val, err := this.data.Get(commons.UInt64(id), readOptions); err != nil {
		if err.Error() == levelDBNotFound {
			return nil, nil
		} else {
			return nil, protocol.ErrorDB(err)
		}
	} else {
		account := &api.Account{}
		if err := json.Unmarshal(val, account); err != nil {
			return nil, protocol.ErrorDB(err)
		}
		return account, nil
	}
}

func (this *AccountServer) Search(gl *registry.GlobalLoading, search *api.Search) (*api.SearchResult, *protocol.TenuredError) {
	return &api.SearchResult{}, nil
}

func (this *AccountServer) Start() (err error) {
	if err = os.MkdirAll(this.dataPath, 0755); err != nil {
		return
	}
	if this.data, err = leveldb.OpenFile(this.dataPath, &opt.Options{Comparer: comparer.DefaultComparer}); err != nil {
		return err
	}
	return nil
}

func (this *AccountServer) Shutdown(interrupt bool) {
	if err := this.data.Close(); err != nil {
		logger.Error("close account error: ", err)
	}
}
