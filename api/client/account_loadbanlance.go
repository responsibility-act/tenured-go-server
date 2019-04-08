package client

import (
	"github.com/ihaiker/tenured-go-server/api"
	"github.com/ihaiker/tenured-go-server/commons/registry"
)

func accountSnowflakeExport(requestCode uint16, obj ...interface{}) uint64 {
	switch requestCode {
	case api.AccountServiceApply:
		return obj[0].(*api.Account).Id
	case api.AccountServiceGet:
		return obj[0].(uint64)
	}
	return 0
}

func HashLoadBalance(serverName string, registration registry.ServiceRegistry) registry.LoadBalance {
	return registry.NewTimedHashLoadBalance(serverName, registration, 100, accountSnowflakeExport)
}
