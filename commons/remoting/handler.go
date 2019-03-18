package remoting

import (
	"github.com/sirupsen/logrus"
)

type RemotingHandler interface {
	//链接事件，当客户端链接时调用,如果客户端返回错误，将直接关闭
	OnChannel(c RemotingChannel) error

	//新消息事件，当客户端发来新的消息时调用
	OnMessage(c RemotingChannel, msg interface{})

	//处理错误事件，当OnMessage抛出未能处理的错误
	OnError(c RemotingChannel, err error, msg interface{})

	//发送心跳包
	OnIdle(c RemotingChannel)

	//关闭事件，当当前客户端关闭连接
	OnClose(c RemotingChannel)
}

type RemotingHandlerFactory func(RemotingChannel, RemotingConfig) RemotingHandler

type HandlerWrapper struct {
}

func (h *HandlerWrapper) OnChannel(c RemotingChannel) error {
	logrus.Debugf("RemotingHandler OnChannel %s", c.RemoteAddr())
	return nil
}
func (h *HandlerWrapper) OnMessage(c RemotingChannel, msg interface{}) {
	logrus.Debugf("RemotingHandler OnMessage %s : msg:%v", c.RemoteAddr(), msg)
}

func (h *HandlerWrapper) OnClose(c RemotingChannel) {
	logrus.Debugf("RemotingHandler OnClose %s ", c.RemoteAddr())
}

func (h *HandlerWrapper) OnError(c RemotingChannel, err error, msg interface{}) {
	logrus.Debugf("RemotingHandler OnError %s : %s ,%v", c.RemoteAddr(), err, msg)
}

func (h *HandlerWrapper) OnIdle(c RemotingChannel) {
	logrus.Debugf("RemotingHandler OnIdle : %s", c.RemoteAddr())
}
