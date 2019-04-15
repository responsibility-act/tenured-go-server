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
		//case api.AccountServiceGetByMobile, api.AccountServiceGetByEmail:
		//	mobileOrEmail := obj[0].(string)
		//	return crc64.Checksum([]byte(mobileOrEmail), crc64.MakeTable(crc64.ECMA))
	}
	return 0
}

func HashLoadBalance(serverName, serverTag string, registration registry.ServiceRegistry) registry.LoadBalance {
	return registry.NewTimedHashLoadBalance(serverName, serverTag, registration, 100, accountSnowflakeExport)
}

func SearchLoadBalance(serverName, serverTag string, registration registry.ServiceRegistry) registry.LoadBalance {
	return nil
}
