package protocol

import (
	"github.com/ihaiker/tenured-go-server/commons/remoting"
)

type TenuredClient struct {
	tenuredService
	Module *AuthHeader
}

func (this *TenuredClient) OnChannel(channel remoting.RemotingChannel) error {
	//request := NewRequest(REQUEST_CODE_ATUH)
	//if err := request.SetHeader(this.Module); err != nil {
	//	return err
	//}
	//resp, err := this.Invoke(channel.RemoteAddr(), request, time.Second*3)
	//if err != nil {
	//	return err
	//} else if !resp.IsSuccess() {
	//	return resp.GetError()
	//}
	//
	//header := &AuthHeader{}
	//if err := resp.GetHeader(header); err != nil {
	//	logrus.Warning("Cannot get the information returned by the server:", err.Error())
	//	return nil
	//} else {
	//	logrus.Info("Get the information returned by the server:", header)
	//}
	return nil
}

func (this *TenuredClient) Start() error {
	if this.Module == nil {
		return ErrorNoModule()
	}
	return this.tenuredService.Start()
}

func NewTenuredClient(config *remoting.RemotingConfig) (*TenuredClient, error) {
	if config == nil {
		config = remoting.DefaultConfig()
	}
	remotingClient := remoting.NewRemotingClient(config)
	remotingClient.SetCoder(&tenuredCoder{config: config})
	client := &TenuredClient{
		tenuredService: tenuredService{
			remoting:         remotingClient,
			responseTables:   map[uint32]*responseTableBlock{},
			commandProcesser: map[uint16]*tenuredCommandRunner{},
		},
	}
	remotingClient.SetHandler(client)
	return client, nil
}
