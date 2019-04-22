package protocol

import (
	"github.com/ihaiker/tenured-go-server/commons/logs"
	"github.com/sirupsen/logrus"
)

const _LoggerName = "remoting"

var logger *logrus.Logger

func init() {
	logger = logs.GetLogger(_LoggerName)
}
