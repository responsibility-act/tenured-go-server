package leveldb

import (
	"github.com/ihaiker/tenured-go-server/api"
	"github.com/ihaiker/tenured-go-server/commons/registry"
	"github.com/ihaiker/tenured-go-server/commons/registry/load_balance"
)

func accountSnowflakeExport(requestCode uint16, obj ...interface{}) uint64 {
	switch requestCode {
	case api.AccountServiceApply:
		return obj[0].(*api.Account).Id
	case api.AccountServiceGet:
		return obj[0].(uint64)
		//case api.AccountServiceGetByMobile, api.AccountServiceGetByEmail:
		//	mobileOrEmail := obj[0].(string)
		//	return crc64.Checksum([]byte(mobileOrEmail), crc64.MakeTable(crc64.ECMA))
	}
	return 0
}

func HashLoadBalance(serverName, serverTag string, registration registry.ServiceRegistry) load_balance.LoadBalance {
	return load_balance.NewTimedHashLoadBalance(serverName, serverTag, registration, 100, accountSnowflakeExport)
}

func SearchLoadBalance(serverName, serverTag string, registration registry.ServiceRegistry) load_balance.LoadBalance {
	return load_balance.NewHashLoadBalance(serverName, serverTag, registration, 100)
}

func NewLoadBalance(serverName string, registration registry.ServiceRegistry) load_balance.LoadBalance {
	lbm := load_balance.NewLoadBalanceManager(nil)

	timedHashLoadBalance := HashLoadBalance(serverName, "account", registration)
	for requestCode := api.AccountServiceApply; requestCode < api.AccountServiceSearchApp; requestCode++ {
		lbm.AddLoadBalance(requestCode, timedHashLoadBalance)
	}

	searchLoadBalance := SearchLoadBalance(serverName, "search", registration)
	for requestCode := api.SearchServicePut; requestCode < api.SearchServiceRemove; requestCode++ {
		lbm.AddLoadBalance(requestCode, searchLoadBalance)
	}

	return lbm
}
