package protocol

import (
	"github.com/ihaiker/tenured-go-server/commons"
	"github.com/ihaiker/tenured-go-server/commons/registry"
	"github.com/ihaiker/tenured-go-server/commons/remoting"
	"time"
)

type TenuredClientInvoke struct {
	ServerName  string
	registry    registry.ServiceRegistry
	loadBalance registry.LoadBalance
	client      *TenuredClient
}

func (this *TenuredClientInvoke) selectOne(zone interface{}) (instance registry.ServerInstance, retErr *TenuredError) {
	ss, err := this.loadBalance.Select(this.ServerName, zone, this.registry)
	if err != nil {
		return instance, ErrorHandler(err)
	} else if len(ss) == 0 || !registry.IsOK(ss[0]) {
		return instance, ErrorHandler(commons.Error("no active server"))
	}
	return ss[0], nil
}

func (this *TenuredClientInvoke) Invoke(
	code uint16, header interface{}, body []byte, timeout time.Duration, respHeader interface{},
) *TenuredError {
	serverInstance, err := this.selectOne(header)
	if err != nil {
		return err
	}
	request := NewRequest(code)
	if header != nil {
		_ = request.SetHeader(header)
	}
	if body != nil {
		request.Body = body
	}
	response, invokeErr := this.client.Invoke(serverInstance.Address, request, timeout)
	if invokeErr != nil {
		return ConvertError(invokeErr)
	}
	if !response.IsSuccess() {
		return ConvertError(response.GetError())
	}
	if err := response.GetHeader(respHeader); err != nil {
		return ConvertError(err)
	}
	return nil
}

func (this *TenuredClientInvoke) initTenuredClient() (err error) {
	if this.client, err = NewTenuredClient(remoting.DefaultConfig()); err != nil {
		return
	}
	this.client.AuthHeader = &AuthHeader{}
	return this.client.Start()
}

func (this *TenuredClientInvoke) Start() (err error) {
	if err = this.initTenuredClient(); err != nil {
		return
	}
	return nil
}

func (this *TenuredClientInvoke) Shutdown(interrupt bool) {
	this.client.Shutdown(interrupt)
}

func NewClient(serverName string, registry registry.ServiceRegistry, loadBalance registry.LoadBalance) *TenuredClientInvoke {
	serverClient := &TenuredClientInvoke{
		ServerName:  serverName,
		registry:    registry,
		loadBalance: loadBalance,
	}
	return serverClient
}
