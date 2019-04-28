package protocol

import (
	"github.com/ihaiker/tenured-go-server/commons/c8tmap"
	"github.com/ihaiker/tenured-go-server/commons/remoting"
)

type SessionManager interface {
	OnConnect(channel remoting.RemotingChannel)
	OnClose(channel remoting.RemotingChannel)

	Size() int
	Get(address string) remoting.RemotingChannel

	Filter(func(remoting.RemotingChannel) bool) []remoting.RemotingChannel
}

type MapSessionManager struct {
	sessions c8tmap.ConcurrentMap
}

func NewMapSessionManager() SessionManager {
	return &MapSessionManager{sessions: c8tmap.New()}
}

func (this *MapSessionManager) OnConnect(channel remoting.RemotingChannel) {
	logger.Debug("new channel ", channel.RemoteAddr())
	this.sessions.Set(channel.RemoteAddr(), channel)
}

func (this *MapSessionManager) OnClose(channel remoting.RemotingChannel) {
	logger.Debug("close channel ", channel.RemoteAddr())
	this.sessions.Remove(channel.RemoteAddr())
}

func (this *MapSessionManager) Size() int {
	return this.sessions.Count()
}

func (this *MapSessionManager) Filter(filter func(remoting.RemotingChannel) bool) []remoting.RemotingChannel {
	channels := make([]remoting.RemotingChannel, 0)
	this.sessions.IterCb(func(key interface{}, v interface{}) {
		if filter(v.(remoting.RemotingChannel)) {
			channels = append(channels, v.(remoting.RemotingChannel))
		}
	})
	return channels
}

func (this *MapSessionManager) Get(address string) remoting.RemotingChannel {
	if ch, has := this.sessions.Get(address); has {
		return ch.(remoting.RemotingChannel)
	}
	return nil
}
