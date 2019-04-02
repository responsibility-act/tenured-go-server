package store

import (
	"github.com/golang/leveldb"
	"github.com/golang/leveldb/db"
	"github.com/ihaiker/tenured-go-server/api"
	"github.com/ihaiker/tenured-go-server/api/command"
	"github.com/ihaiker/tenured-go-server/commons/executors"
	"github.com/ihaiker/tenured-go-server/commons/protocol"
	"os"
)

type AccountServer struct {
	accountDB *leveldb.DB

	executorSet []executors.ExecutorService
}

func (this *AccountServer) Apply(account *command.Account) (*command.Account, *protocol.TenuredError) {

	return account, nil
}

func (this *AccountServer) Start() error {
	return nil
}

func (this *AccountServer) Shutdown(interrupt bool) {
	//close executor
	for _, v := range this.executorSet {
		v.Shutdown(interrupt)
	}
	//close level db
	if err := this.accountDB.Close(); err != nil {
		logger.Error("close account error: ", err)
	} else {
		logger.Debug("close account db")
	}
}

func NewAccountServer(config *storeConfig, server *protocol.TenuredServer) (accountServer *AccountServer, err error) {
	accountServer = &AccountServer{
		executorSet: make([]executors.ExecutorService, 0),
	}

	accountDBPath := config.Data + "/store/account"

	if err = os.MkdirAll(accountDBPath, 0755); err != nil {
		return
	}

	if accountLevelDB, err := leveldb.Open(accountDBPath, &db.Options{}); err != nil {
		return nil, err
	} else {
		accountServer.accountDB = accountLevelDB
	}

	invoke := protocol.NewInvoke(server, accountServer)

	executor := executors.NewFixedExecutorService(
		config.Executors.Get("accountSize", 10),
		config.Executors.Get("accountBuffer", 1000),
	)
	if err = invoke.Invoke(api.AccountServiceApply, "Apply", executor); err != nil {
		return
	}
	accountServer.executorSet = append(accountServer.executorSet, executor)
	return
}
