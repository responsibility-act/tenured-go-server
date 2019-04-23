package leveldb

import (
	"github.com/ihaiker/tenured-go-server/api"
	"github.com/ihaiker/tenured-go-server/registry"
	"github.com/ihaiker/tenured-go-server/registry/load_balance"
)

func AccountLoadBalance(serverName, serverTag string, reg registry.ServiceRegistry) load_balance.LoadBalance {
	return load_balance.NewTimedHashLoadBalance(serverName, serverTag, reg, 100, func(requestCode uint16, obj ...interface{}) uint64 {
		switch requestCode {
		case api.AccountServiceApply:
			return obj[0].(*api.Account).Id
		case api.AccountServiceGet:
			return obj[0].(uint64)
		}
		return 0
	})
}

func SearchLoadBalance(serverName, serverTag string, reg registry.ServiceRegistry) load_balance.LoadBalance {
	return load_balance.NewHashLoadBalance(serverName, serverTag, reg, 100)
}

func UserLoadBalance(serverName, serverTag string, reg registry.ServiceRegistry) load_balance.LoadBalance {
	return load_balance.NewRoundLoadBalance(serverName, serverTag, reg)
}

func NewLoadBalance(serverName string, reg registry.ServiceRegistry) load_balance.LoadBalance {
	lbm := load_balance.NewLoadBalanceManager(nil)

	//account
	{
		timedHashLoadBalance := AccountLoadBalance(serverName, api.StoreAccount, reg)
		for requestCode := api.AccountServiceRange.Min; requestCode < api.AccountServiceRange.Max; requestCode++ {
			lbm.AddLoadBalance(requestCode, timedHashLoadBalance)
		}
		round := load_balance.NewRoundLoadBalance(serverName, api.StoreAccount, reg)
		lbm.AddLoadBalance(api.AccountServiceGetByMobile, round)
		lbm.AddLoadBalance(api.AccountServiceGetByEmail, round)
	}

	//search
	{
		searchLoadBalance := SearchLoadBalance(serverName, api.StoreSearch, reg)
		for requestCode := api.SearchServiceRange.Min; requestCode < api.SearchServiceRange.Max; requestCode++ {
			lbm.AddLoadBalance(requestCode, searchLoadBalance)
		}
	}

	//user
	{
		userLoadBalance := UserLoadBalance(serverName, api.StoreUser, reg)
		for requestCode := api.UserServiceRange.Min; requestCode < api.UserServiceRange.Max; requestCode++ {
			lbm.AddLoadBalance(requestCode, userLoadBalance)
		}
		round := load_balance.NewRoundLoadBalance(serverName, api.StoreUser, reg)
		lbm.AddLoadBalance(api.UserServiceGetByTenantUserId, round)
	}

	//snowflake
	{
		lbm.AddLoadBalance(api.ClusterIdServiceGet, load_balance.NewRoundLoadBalance(serverName, api.StoreClusterId, reg))
	}
	return lbm
}
