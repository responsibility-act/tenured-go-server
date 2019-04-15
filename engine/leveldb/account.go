package leveldb

import (
	"encoding/json"
	"fmt"
	"github.com/ihaiker/tenured-go-server/api"
	"github.com/ihaiker/tenured-go-server/commons/protocol"
	"github.com/ihaiker/tenured-go-server/commons/registry"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/comparer"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	MIN_ID = 10000000000000000
	MAX_ID = 99999999999999999
)

func accountKey(id uint64) []byte {
	return []byte(fmt.Sprintf("A:%d", id))
}
func statusKey(id uint64) []byte {
	return []byte(fmt.Sprintf("S:%d", MAX_ID-id))
}

type AccountServer struct {
	dataPath string
	data     *leveldb.DB
}

func NewAccountServer(dataPath string) (*AccountServer, error) {
	accountServer := &AccountServer{
		dataPath: dataPath + "/store/account",
	}
	return accountServer, nil
}

func (this *AccountServer) Apply(account *api.Account) *protocol.TenuredError {
	logger.Debug("申请用户：", account)

	if _, err := this.Get(account.Id); err != nil && err != api.ErrAccountNotExists {
		return api.ErrAccountExists
	}

	account.Status = api.AccountStatusApply
	bs, _ := json.Marshal(account)

	//保存用户信息
	batch := &leveldb.Batch{}
	batch.Put(accountKey(account.Id), bs)
	batch.Put(statusKey(account.Id), []byte(api.AccountStatusApply))
	if err := this.data.Write(batch, writeOptions); err != nil {
		return protocol.ErrorDB(err)
	}
	return nil
}

func (this *AccountServer) Get(id uint64) (*api.Account, *protocol.TenuredError) {
	logger.Debug("获取用户: ", id)
	if val, err := this.data.Get(accountKey(id), readOptions); err != nil {
		if err.Error() == levelDBNotFound {
			return nil, api.ErrAccountNotExists
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
	logger.Debug("搜索：", search)

	if sn, err := this.data.GetSnapshot(); err != nil {
		return nil, protocol.ConvertError(err)
	} else {
		sr := &api.SearchResult{Accounts: make([]*api.Account, 0)}

		var startKey []byte
		if search.StartId == 0 {
			startKey = statusKey(MAX_ID - MIN_ID)
		} else {
			startKey = statusKey(search.StartId)
		}

		resultSize := 0
		it := sn.NewIterator(&util.Range{Start: startKey}, readOptions)
		defer it.Release()
		for it.Next() && resultSize < search.Limit {
			key := string(it.Key())
			if !strings.HasPrefix(key, "S:") {
				break
			}
			if "" != string(search.Status) &&
				search.Status != api.AccountStatus(string(it.Value())) {
				continue
			}
			id, _ := strconv.ParseUint(key[2:], 10, 64)
			if search.StartId != 0 && search.StartId == MAX_ID-id {
				continue
			}
			if account, err := this.Get(MAX_ID - id); err != nil {
				return nil, err
			} else {
				resultSize++
				sr.Accounts = append(sr.Accounts, account)
			}
		}
		return sr, nil
	}
}

func (this *AccountServer) Check(checkAccount *api.CheckAccount) *protocol.TenuredError {
	if ac, err := this.Get(checkAccount.Id); err != nil {
		return err
	} else {
		batch := &leveldb.Batch{}

		statusKey := statusKey(ac.Id)

		switch checkAccount.Status {
		case api.AccountStatusOK:
			{
				batch.Delete(statusKey)
			}
		case api.AccountStatusReturn, api.AccountStatusDeny, api.AccountStatusDisable:
			{
				batch.Put(statusKey, []byte(checkAccount.Status))
			}
		}
		ac.Status = checkAccount.Status
		ac.StatusDescription = checkAccount.StatusDescription
		ac.StatusTime = time.Now().Format("2006-01-02 15:04:05")

		bs, _ := json.Marshal(ac)
		batch.Put(accountKey(ac.Id), bs)
		if err := this.data.Write(batch, writeOptions); err != nil {
			return protocol.ErrorDB(err)
		}
		return nil
	}
}

func (this *AccountServer) Start() (err error) {
	logger.Debug("start accout store.")
	if err = os.MkdirAll(this.dataPath, 0755); err != nil {
		logger.Error("start account store error: ", err)
		return
	}
	if this.data, err = leveldb.OpenFile(this.dataPath, &opt.Options{Comparer: comparer.DefaultComparer}); err != nil {
		logger.Error("start account store error: ", err)
		return err
	}
	return nil
}

func (this *AccountServer) Shutdown(interrupt bool) {
	if err := this.data.Close(); err != nil {
		logger.Error("close account error: ", err)
	}
}
