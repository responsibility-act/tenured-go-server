package leveldb

import (
	"fmt"
	"github.com/ihaiker/tenured-go-server/api"
	"github.com/ihaiker/tenured-go-server/commons/snowflake"
	"github.com/stretchr/testify/assert"
	"testing"
)

var dataDir = "/data/tenured"

var accountService = NewAccountServer(dataDir)
var ID uint64 = 19416244780269568
var sf = snowflake.NewSnowflake(snowflake.Settings{MachineID: 0})

func init() {
	if err := accountService.Start(); err != nil {
		panic(err)
	}
}

func TestAccountServer_Apply(t *testing.T) {
	account := &api.Account{}
	account.Id, _ = sf.NextID()
	account.Name = fmt.Sprintf("名称：%d", account.Id)

	err := accountService.Apply(account)
	assert.Nil(t, err)
}

func BenchmarkAccountServer_Apply(t *testing.B) {
	for i := 0; i < t.N; i++ {
		account := &api.Account{}
		account.Id, _ = sf.NextID()
		account.Name = fmt.Sprintf("名称: %d", i)

		err := accountService.Apply(account)
		assert.Nil(t, err)
	}
}

func TestAccountServer_Get(t *testing.T) {
	ac, err := accountService.Get(ID)
	assert.Nil(t, err)
	t.Log(ac)
}

func TestAccountServer_Check(t *testing.T) {
	check := &api.CheckAccount{
		Id:                19607670029811712,
		Status:            api.AccountStatusReturn,
		StatusDescription: "您的照片信息有误，请重新上传！",
	}
	err := accountService.Check(check)
	t.Log(err)
}

func TestAccountServer_Search(t *testing.T) {
	search := new(api.Search)
	search.Limit = 10
	//search.Status = api.AccountStatusDisable
	idx := 0
	for {
		rs, err := accountService.Search(nil, search)
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
