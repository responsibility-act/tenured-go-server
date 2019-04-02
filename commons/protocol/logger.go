package protocol

import (
	"github.com/ihaiker/tenured-go-server/commons/logs"
	"github.com/sirupsen/logrus"
)

const _LoggerName = "remoting"

var protocolLogger *logrus.Logger

func init() {
	protocolLogger = logs.GetLogger(_LoggerName)
}

func logger() *logrus.Entry {
	return protocolLogger.WithField("agent", _LoggerName)
}
