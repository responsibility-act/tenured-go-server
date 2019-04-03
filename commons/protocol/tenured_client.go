package protocol

import (
	"github.com/ihaiker/tenured-go-server/commons/remoting"
	"time"
)

type TenuredClient struct {
	tenuredService
	*AuthHeader
}

func (this *TenuredClient) OnChannel(channel remoting.RemotingChannel) error {
	logger.Debug("send auth code:", channel.RemoteAddr())
	request := NewRequest(REQUEST_CODE_ATUH)
	if err := request.SetHeader(this.AuthHeader); err != nil {
		return err
	}
	resp, err := this.Invoke(channel.RemoteAddr(), request, time.Second*3)
	if err != nil {
		logger.Debug("send auth error:", err)
		return err
	} else if !resp.IsSuccess() {
		err = resp.GetError()
		logger.Debug("send auth error:", err)
		return err
	}

	header := &AuthHeader{}
	if err := resp.GetHeader(header); err != nil {
		logger.Warning("Cannot get the information returned by the server: ", err.Error())
		return nil
	} else {
		logger.Info("Get the information returned by the server:", header)
	}
	return nil
}

func (this *TenuredClient) Start() error {
	if this.AuthHeader == nil {
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
