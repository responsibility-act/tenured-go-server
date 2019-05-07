package leveldb

import (
	"encoding/json"
	"fmt"
	"github.com/ihaiker/tenured-go-server/api/client"
	"github.com/ihaiker/tenured-go-server/protocol"
	"github.com/ihaiker/tenured-go-server/registry"
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

func mobileKey(mobile string) string {
	return fmt.Sprintf("Mobile:%s", mobile)
}
func emailKey(email string) string {
	return fmt.Sprintf("Email:%s", email)
}

type AccountServer struct {
	storeName string
	dataPath  string
	data      *leveldb.DB

	reg            registry.ServiceRegistry
	loadBalance    load_balance.LoadBalance
	search         api.SearchService
	accountService api.AccountService
}

func NewAccountServer(storeName, dataPath string) (*AccountServer, error) {
	accountServer := &AccountServer{
		storeName: storeName,
		dataPath:  dataPath + "/store/account",
	}
	return accountServer, nil
}

func (this *AccountServer) Apply(account *api.Account) *protocol.TenuredError {
	logger.Debug("申请用户：", account)

	if _, err := this.Get(account.Id); err != api.ErrAccountNotExists {
		return api.ErrAccountExists
	}

	idbyte := []byte(fmt.Sprintf("%d", account.Id))
	if account.Email != "" {
		if err := this.search.Put(emailKey(account.Email), idbyte); commons.NotNil(err) {
			if err.Code() == api.ErrSearchExists.Code() {
				return api.ErrEmailRegistered
			} else {
				return err
			}
		}
	}
	if account.Mobile != "" {
		if err := this.search.Put(mobileKey(account.Mobile), idbyte); commons.NotNil(err) {
			if err.Code() == api.ErrSearchExists.Code() {
				return api.ErrMobileRegistered
			} else {
				return err
			}
		}
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
		return nil, notFound(err, api.ErrAccountNotExists)
	} else {
		account := &api.Account{}
		_ = json.Unmarshal(val, account)
		return account, nil
	}
}

//根据手机号获取用户信息
func (this *AccountServer) GetByMobile(mobile string) (*api.Account, *protocol.TenuredError) {
	logger.Debug("获取账户 mobile: ", mobile)
	key := mobileKey(mobile)
	accountId := uint64(0)
	if val, err := this.search.Get(string(key)); err != nil {
		if err.Code() == api.ErrSearchNotExists.Code() {
			return nil, api.ErrAccountNotExists
		}
		return nil, err
	} else {
		accountId, _ = strconv.ParseUint(string(val), 10, 64)
	}

	return this.accountService.Get(accountId)
}

//根据邮箱获取用户信息
func (this *AccountServer) GetByEmail(email string) (*api.Account, *protocol.TenuredError) {
	logger.Debug("获取账户 Email: ", email)
	key := emailKey(email)
	accountId := uint64(0)
	if val, err := this.search.Get(string(key)); err != nil {
		if err.Code() == api.ErrSearchNotExists.Code() {
			return nil, api.ErrAccountNotExists
		}
		return nil, err
	} else {
		accountId, _ = strconv.ParseUint(string(val), 10, 64)
	}
	return this.accountService.Get(accountId)
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
	app.CreateTime = time.Now().Format("2006-01-02 15:04:05")

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

func (this *AccountServer) SetRegistry(serviceRegistry registry.ServiceRegistry) {
	this.reg = serviceRegistry
}

func (this *AccountServer) Start() (err error) {
	this.loadBalance = NewLoadBalance(this.storeName, this.reg)

	this.search = client.NewSearchServiceClient(this.loadBalance)
	this.accountService = client.NewAccountServiceClient(this.loadBalance)

	logger.Debug("start account store.")
	if err = os.MkdirAll(this.dataPath, 0755); err != nil {
		logger.Error("start account store error: ", err)
		return
	}
	if this.data, err = leveldb.OpenFile(this.dataPath, &opt.Options{Comparer: comparer.DefaultComparer}); err != nil {
		logger.Error("start account store error: ", err)
		return err
	}
	if err = commons.StartIfService(this.search); err != nil {
		return
	}
	if err = commons.StartIfService(this.accountService); err != nil {
		return
	}
	return commons.StartIfService(this.loadBalance)
}

func (this *AccountServer) Shutdown(interrupt bool) {
	if err := this.data.Close(); err != nil {
		logger.Error("close account error: ", err)
	}
	commons.ShutdownIfService(this.search, interrupt)
	commons.ShutdownIfService(this.accountService, interrupt)
	commons.ShutdownIfService(this.loadBalance, interrupt)
}
