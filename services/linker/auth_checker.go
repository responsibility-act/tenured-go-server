package linker

import (
	"github.com/ihaiker/tenured-go-server/api"
	"github.com/ihaiker/tenured-go-server/api/client"
	"github.com/ihaiker/tenured-go-server/commons/remoting"
	"github.com/ihaiker/tenured-go-server/protocol"
	"github.com/ihaiker/tenured-go-server/registry/load_balance"
)

type Auth struct {
	//客户端IP地址
	Address string `json:"address"`

	Token string `json:"token"`

	//用户所属账户ID
	AccountId uint64 `json:"accountId"`

	//用户所属AppId
	AppId uint64 `json:"appId"`

	//云用户ID
	CloudId uint64 `json:"cloudId"`
}

type LinkerAuthChecker struct {
	serverAddress string
	userServer    api.UserService
}

func NewLinkerAuthChecker(serverAddress string, loadBalance load_balance.LoadBalance) (*LinkerAuthChecker, error) {
	s := &LinkerAuthChecker{
		serverAddress: serverAddress,
		userServer:    client.NewUserServiceClient(loadBalance),
	}
	return s, nil
}

func (this *LinkerAuthChecker) Auth(channel remoting.RemotingChannel, command *protocol.TenuredCommand) *protocol.TenuredError {
	auth := new(Auth)
	if err := command.GetHeader(auth); err != nil {
		return ErrAuth
	}
	logger.Info("用户认证：", auth)

	if token, err := this.userServer.GetToken(auth.AccountId, auth.AppId, auth.CloudId); err != nil {
		return err
	} else if token.Token != auth.Token || token.Linker != this.serverAddress {
		logger.Info("用户非法连接：", auth)
		return ErrAuth
	}
	channel.Attributes()["auth"] = auth
	return nil
}

func (this *LinkerAuthChecker) IsAuthed(channel remoting.RemotingChannel) bool {
	attr := channel.Attributes()
	_, has := attr["auth"]
	return has
}
