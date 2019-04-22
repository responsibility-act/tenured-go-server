package protocol

import (
	"github.com/ihaiker/tenured-go-server/commons/remoting"
)

const auth_attributes_name = "auth_token"

type TenuredAuthChecker interface {
	Auth(channel remoting.RemotingChannel, command *TenuredCommand) error
	IsAuthed(channel remoting.RemotingChannel) bool
}

type ModuleAuthChecker struct {
}

func (this *ModuleAuthChecker) Auth(channel remoting.RemotingChannel, command *TenuredCommand) error {
	header := &AuthHeader{}
	if err := command.GetHeader(header); err != nil {
		return err
	} else {
		channel.Attributes()[auth_attributes_name] = true
	}

	return nil
}

func (this *ModuleAuthChecker) IsAuthed(channel remoting.RemotingChannel) bool {
	attrs := channel.Attributes()
	if attrs == nil {
		return false
	}
	_, has := attrs[auth_attributes_name]
	return has
}
