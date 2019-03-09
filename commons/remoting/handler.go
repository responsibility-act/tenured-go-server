package remoting

import (
	"github.com/sirupsen/logrus"
)

type Handler interface {
	//链接事件，当客户端链接时调用
	OnChannel(c Channel)

	//新消息事件，当客户端发来新的消息时调用
	OnMessage(c Channel, msg interface{})

	//处理错误事件，当OnMessage抛出未能处理的错误
	OnError(c Channel, err error, msg interface{})

	//发送心跳包
	OnIdle(c Channel)

	//关闭事件，当当前客户端关闭连接
	OnClose(c Channel)
}

type HandlerFactory func(Channel) Handler

type HandlerWrapper struct {
}

func (h *HandlerWrapper) OnChannel(c Channel) {
	logrus.Debugf("Handler OnChannel %s", c.RemoteAddr())
}
func (h *HandlerWrapper) OnMessage(c Channel, msg interface{}) {
	logrus.Debugf("Handler OnMessage %s : msg:%s", c.RemoteAddr(), msg)
}

func (h *HandlerWrapper) OnClose(c Channel) {
	logrus.Debugf("Handler OnClose %s ", c.RemoteAddr())
}

func (h *HandlerWrapper) OnError(c Channel, err error, msg interface{}) {
	logrus.Debugf("Handler OnError %s : %s ,%s", c.RemoteAddr(), err, msg)
}

func (h *HandlerWrapper) OnIdle(c Channel) {
	logrus.Debugf("Handler OnIdle : %s", c.RemoteAddr())
}
