package dao

import (
	"github.com/golang/leveldb"
	"github.com/golang/leveldb/db"
	"github.com/ihaiker/tenured-go-server/api"
	"github.com/ihaiker/tenured-go-server/commons/protocol"
	"os"
)

type AccountServer struct {
	dataPath string
	data     *leveldb.DB
}

func NewAccountServer(dataPath string) *AccountServer {
	return &AccountServer{dataPath: dataPath}
}

func (this *AccountServer) Apply(account *api.Account) (*api.Account, *protocol.TenuredError) {
	b := leveldb.Batch{}
	b.Set([]byte(account.ID+"-ID"), []byte(account.ID))
	b.Set([]byte(account.ID+"-NAME"), []byte(account.Name))
	err := this.data.Apply(b, &db.WriteOptions{Sync: true})
	if err != nil {
		return nil, protocol.ErrorHandler(err)
	}
	return account, nil
}

func (this *AccountServer) GetById(id string) (*api.Account, *protocol.TenuredError) {
	return nil, nil
}

func (this *AccountServer) Start() (err error) {
	if err = os.MkdirAll(this.dataPath, 0755); err != nil {
		return
	}
	if this.data, err = leveldb.Open(this.dataPath, &db.Options{}); err != nil {
		return err
	}
	return nil
}

func (this *AccountServer) Shutdown(interrupt bool) {
	if err := this.data.Close(); err != nil {
		logger.Error("close account error: ", err)
	}
}
