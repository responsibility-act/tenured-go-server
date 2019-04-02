package remoting

import (
	"github.com/ihaiker/tenured-go-server/commons/logs"
	"github.com/sirupsen/logrus"
)

var remotingLogger *logrus.Logger

func init() {
	remotingLogger = logs.GetLogger("remoting")
}

func logger() *logrus.Entry {
	return remotingLogger.WithField("agent", "remoting")
}
