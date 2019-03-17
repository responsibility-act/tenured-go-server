package protocol

import (
	"github.com/ihaiker/tenured-go-server/commons/remoting"
	"time"
)

type TenuredChannel struct {
	Headers *AuthHeader
	channel remoting.RemotingChannel
}

func (this *TenuredChannel) RemoteAddr() string {
	return this.channel.RemoteAddr()
}
func (this *TenuredChannel) ChannelAttributes() map[string]string {
	return this.channel.ChannelAttributes()
}
func (this *TenuredChannel) Write(msg interface{}, timeout time.Duration) error {
	return this.channel.Write(msg, timeout)
}
func (this *TenuredChannel) AsyncWrite(msg interface{}, timeout time.Duration, callback func(error)) {
	this.channel.AsyncWrite(msg, timeout, callback)
}
func (this *TenuredChannel) Close() {
	this.channel.Close()
}

type TenuredRemotingChannelTransfer struct {
}

func (this *TenuredRemotingChannelTransfer) Transform(channel remoting.RemotingChannel) remoting.RemotingChannel {
	return &TenuredChannel{channel: channel, Headers: nil}
}
