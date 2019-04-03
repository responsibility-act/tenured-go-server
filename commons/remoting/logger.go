package remoting

import (
	"github.com/ihaiker/tenured-go-server/commons/logs"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

func init() {
	logger = logs.GetLogger("remoting")
}
