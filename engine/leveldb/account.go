package leveldb

import (
	"encoding/json"
	"fmt"
	"github.com/ihaiker/tenured-go-server/protocol"
	"github.com/ihaiker/tenured-go-server/registry/load_balance"

	"github.com/ihaiker/tenured-go-server/api"
	"github.com/ihaiker/tenured-go-server/commons"
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

func appKey(accountId, appId uint64) []byte {
	return []byte(fmt.Sprintf("P:%d,%d", accountId, appId))
}
func appStatusKey(accountId, appId uint64) []byte {
	return []byte(fmt.Sprintf("T:%d,%d", MAX_ID-accountId, MAX_ID-appId))
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

	if _, err := this.Get(account.Id); err != api.ErrAccountNotExists {
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

func (this *AccountServer) Search(gl *load_balance.GlobalLoading, search *api.Search) (*api.SearchResult, *protocol.TenuredError) {
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

//添加APP
func (this *AccountServer) ApplyApp(app *api.App) *protocol.TenuredError {
	logger.Debug("申请App：", app)

	if _, err := this.GetApp(app.AccountId, app.Id); err != api.ErrAccountAppNotExists {
		return api.ErrAccountAppExists
	}

	app.Status = api.AccountStatusApply

	bs, _ := json.Marshal(app)
	batch := &leveldb.Batch{}
	batch.Put(appKey(app.AccountId, app.Id), bs)
	batch.Put(appStatusKey(app.AccountId, app.Id), []byte(api.AccountStatusApply))

	if err := this.data.Write(batch, writeOptions); err != nil {
		return protocol.ErrorDB(err)
	}
	return nil
}

//搜索账户APP
func (this *AccountServer) SearchApp(searchApp *api.SearchApp) (*api.SearchAppResult, *protocol.TenuredError) {
	logger.Debug("搜索：", searchApp)

	if sn, err := this.data.GetSnapshot(); err != nil {
		return nil, protocol.ConvertError(err)
	} else {
		sr := &api.SearchAppResult{SearchApps: make([]*api.App, 0)}

		var startKey []byte
		if searchApp.StartId == 0 {
			startKey = appStatusKey(searchApp.AccountId, MAX_ID-MIN_ID)
		} else {
			startKey = appStatusKey(searchApp.AccountId, searchApp.StartId)
		}

		resultSize := 0
		it := sn.NewIterator(&util.Range{Start: startKey}, readOptions)
		defer it.Release()
		for it.Next() && resultSize < searchApp.Limit {
			key := string(it.Key())
			if !strings.HasPrefix(key, "T:") {
				break
			}
			if "" != string(searchApp.Status) &&
				searchApp.Status != api.AccountStatus(string(it.Value())) {
				continue
			}
			accountId, appId, _ := commons.SplitToUint2(key[2:], 10, 64)
			if searchApp.StartId != 0 && searchApp.StartId == MAX_ID-appId {
				continue
			}
			if app, err := this.GetApp(MAX_ID-accountId, MAX_ID-appId); err != nil {
				return nil, err
			} else {
				resultSize++
				sr.SearchApps = append(sr.SearchApps, app)
			}
		}
		return sr, nil
	}
}

func (this *AccountServer) GetApp(accountId uint64, appId uint64) (*api.App, *protocol.TenuredError) {
	key := appKey(accountId, appId)
	if val, err := this.data.Get(key, readOptions); err != nil {
		if err.Error() == levelDBNotFound {
			return nil, api.ErrAccountAppNotExists
		} else {
			return nil, protocol.ErrorDB(err)
		}
	} else {
		app := &api.App{}
		if err := json.Unmarshal(val, app); err != nil {
			return nil, protocol.ErrorDB(err)
		}
		return app, nil
	}
}

//审核APP
func (this *AccountServer) CheckApp(checkAccountApp *api.CheckAccountApp) *protocol.TenuredError {
	if ac, err := this.GetApp(checkAccountApp.AccountId, checkAccountApp.AppId); err != nil {
		return err
	} else {
		batch := &leveldb.Batch{}

		statusKey := appStatusKey(ac.AccountId, ac.Id)

		switch checkAccountApp.Status {
		case api.AccountStatusOK:
			{
				batch.Delete(statusKey)
			}
		case api.AccountStatusReturn, api.AccountStatusDeny, api.AccountStatusDisable:
			{
				batch.Put(statusKey, []byte(checkAccountApp.Status))
			}
		}
		ac.Status = checkAccountApp.Status
		ac.StatusDescription = checkAccountApp.StatusDescription
		ac.StatusTime = time.Now().Format("2006-01-02 15:04:05")

		bs, _ := json.Marshal(ac)
		batch.Put(appKey(ac.AccountId, ac.Id), bs)
		if err := this.data.Write(batch, writeOptions); err != nil {
			return protocol.ErrorDB(err)
		}
		return nil
	}

}

func (this *AccountServer) Start() (err error) {
	logger.Debug("start account store.")
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
