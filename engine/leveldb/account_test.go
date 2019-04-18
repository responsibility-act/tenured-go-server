package leveldb

import (
	"fmt"
	"github.com/ihaiker/tenured-go-server/api"
	"github.com/ihaiker/tenured-go-server/commons/snowflake"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var ID uint64 = 19416244780269568
var sf = snowflake.NewSnowflake(snowflake.Settings{MachineID: 0})

func accountService() *AccountServer {
	var dataDir = "/data/tenured"

	if accountService, err := NewAccountServer(dataDir); err != nil {
		panic(err)
	} else if err := accountService.Start(); err != nil {
		panic(err)
	} else {
		return accountService
	}
}

func TestAccountServer_Apply(t *testing.T) {
	account := &api.Account{}
	account.Id, _ = sf.NextID()
	account.Name = fmt.Sprintf("名称：%d", account.Id)

	err := accountService().Apply(account)
	assert.Nil(t, err)
}

func BenchmarkAccountServer_Apply(t *testing.B) {
	ac := accountService()
	defer ac.Shutdown(true)

	for i := 0; i < t.N; i++ {
		account := &api.Account{}
		account.Id, _ = sf.NextID()
		account.Name = fmt.Sprintf("名称: %d", i)
		err := ac.Apply(account)
		assert.Nil(t, err)
	}
}

func TestAccountServer_Get(t *testing.T) {
	ac, err := accountService().Get(ID)
	assert.Nil(t, err)
	t.Log(ac)
}

func TestAccountServer_Check(t *testing.T) {
	check := &api.CheckAccount{
		Id:                19607670029811712,
		Status:            api.AccountStatusReturn,
		StatusDescription: "您的照片信息有误，请重新上传！",
	}
	err := accountService().Check(check)
	t.Log(err)
}

func TestAccountServer_Search(t *testing.T) {
	search := new(api.Search)
	search.Limit = 10
	//search.Status = api.AccountStatusDisable
	ac := accountService()
	defer ac.Shutdown(true)
	idx := 0
	for {
		rs, err := ac.Search(nil, search)
		assert.Nil(t, err)
		if rs == nil {
			break
		}
		for _, account := range rs.Accounts {
			idx++
			t.Log(idx, account)
			search.StartId = account.Id
		}
		if len(rs.Accounts) < search.Limit {
			break
		}
	}
}

func TestAccountServer_ApplyApp(t *testing.T) {
	app := new(api.App)
	app.AccountId = 1
	app.Id = 2
	app.Name = "appid"
	app.CreateTime = time.Now().Format("2006-01-02 15:04:05")
	err := accountService().ApplyApp(app)
	assert.Nil(t, err)
}

func TestAccountServer_GetApp(t *testing.T) {
	app, err := accountService().GetApp(1, 2)
	assert.Nil(t, err)
	t.Log(app)
}

func TestAccountServer_SearchApp(t *testing.T) {
	search := new(api.SearchApp)
	search.AccountId = 1
	search.Limit = 10

	as := accountService()
	rs, err := as.SearchApp(search)
	assert.Nil(t, err)
	for _, k := range rs.SearchApps {
		t.Log(k)
	}
}

func TestAccountServer_CheckApp(t *testing.T) {
	as := accountService()
	ca := new(api.CheckAccountApp)
	ca.AccountId = 1
	ca.AppId = 2
	ca.Status = api.AccountStatusOK
	err := as.CheckApp(ca)
	assert.Nil(t, err)

	app, err := as.GetApp(1, 2)
	assert.Nil(t, err)
	assert.Equal(t, string(app.Status), string(api.AccountStatusOK))
	t.Log(app)
}
