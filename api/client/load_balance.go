package client

import "github.com/ihaiker/tenured-go-server/commons/registry"

func HashLoadBalance(serverName string, reg registry.ServiceRegistry) registry.LoadBalance {
	return nil
}
