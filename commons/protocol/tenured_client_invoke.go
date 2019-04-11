package protocol

import (
	"github.com/ihaiker/tenured-go-server/commons/registry"
	"github.com/ihaiker/tenured-go-server/commons/remoting"
	"time"
)

type TenuredClientInvoke struct {
	client *TenuredClient
}

func (this *TenuredClientInvoke) Invoke(
	serverInstance *registry.ServerInstance,
	code uint16, header interface{}, body []byte, timeout time.Duration, respHeader interface{},
) ([]byte, *TenuredError) {
	request := NewRequest(code)
	if header != nil {
		if err := request.SetHeader(header); err != nil {
			return nil, ConvertError(err)
		}
	}
	if body != nil {
		request.Body = body
	}
	response, invokeErr := this.client.Invoke(serverInstance.Address, request, timeout)
	if invokeErr != nil {
		return nil, ConvertError(invokeErr)
	}
	if !response.IsSuccess() {
		return nil, response.GetError()
	}
	if respHeader != nil {
		if err := response.GetHeader(respHeader); err != nil {
			return nil, ConvertError(err)
		}
	}
	return response.Body, nil
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

func NewClientInvoke() *TenuredClientInvoke {
	serverClient := &TenuredClientInvoke{}
	return serverClient
}
