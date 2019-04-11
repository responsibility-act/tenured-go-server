package client

import (
	"github.com/ihaiker/tenured-go-server/api"
	"github.com/ihaiker/tenured-go-server/commons"
	"github.com/ihaiker/tenured-go-server/commons/protocol"
	"github.com/ihaiker/tenured-go-server/commons/registry"
	"time"
)

//获取分布式ID
type ClusterIdServiceClient struct {
	*protocol.TenuredClientInvoke
	serverName string
	reg        registry.ServiceRegistry

	roundLB registry.LoadBalance

	serviceManager *commons.ServiceManager
}

func (this *ClusterIdServiceClient) Start() error {
	return this.serviceManager.Start()
}

func (this *ClusterIdServiceClient) Shutdown(interrupt bool) {
	this.serviceManager.Shutdown(interrupt)
}

func (this *ClusterIdServiceClient) Get() (uint64, *protocol.TenuredError) {

	serverInstance, _, err := this.roundLB.Select(api.ClusterIdServiceGet)

	if err != nil || len(serverInstance) == 0 || registry.AllNotOK(serverInstance...) {
		return 0, protocol.ErrorRouter()
	}

	var respBody []byte
	if respBody, err = this.Invoke(serverInstance[0], api.ClusterIdServiceGet, nil, nil, time.Millisecond*3000, nil); !commons.IsNil(err) {
		return 0, protocol.ConvertError(err)
	} else {
		return commons.ToUInt64(respBody), nil
	}

}

func NewClusterIdServiceClient(serverName string, reg registry.ServiceRegistry) (*ClusterIdServiceClient, error) {
	client := &ClusterIdServiceClient{
		TenuredClientInvoke: &protocol.TenuredClientInvoke{},
	}
	client.serverName = serverName
	client.reg = reg
	client.serviceManager = commons.NewServiceManager()
	client.serviceManager.Add(client.TenuredClientInvoke)

	client.roundLB = registry.NewRoundLoadBalance(serverName, reg)
	client.serviceManager.Add(client.roundLB)

	return client, nil
}
