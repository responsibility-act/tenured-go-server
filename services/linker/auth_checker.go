package linker

import (
	"github.com/ihaiker/tenured-go-server/commons/remoting"
	"github.com/ihaiker/tenured-go-server/protocol"
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
	//userServer api.AccountService
}

func NewLinkerAuthChecker() (*LinkerAuthChecker, error) {
	return &LinkerAuthChecker{}, nil
}

func (this *LinkerAuthChecker) Auth(channel remoting.RemotingChannel, command *protocol.TenuredCommand) *protocol.TenuredError {
	auth := new(Auth)
	if err := command.GetHeader(auth); err != nil {
		return ErrAuth
	}
	logger.Info("用户认证：", auth)
	channel.Attributes()["auth"] = auth
	return nil
}

func (this *LinkerAuthChecker) IsAuthed(channel remoting.RemotingChannel) bool {
	attr := channel.Attributes()
	_, has := attr["auth"]
	return has
}
