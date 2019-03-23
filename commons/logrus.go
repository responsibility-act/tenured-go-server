package commons

import (
	"github.com/sirupsen/logrus"
	"strings"
)

type DefaultFieldHook struct{}

func (h *DefaultFieldHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
func (h *DefaultFieldHook) Fire(e *logrus.Entry) error {
	if e.Caller.Function == "github.com/kataras/golog.integrateStdLogger.func1" {
		e.Caller.Function = "iris"
		e.Caller.File = "iris"
		return nil
	}

	idx := strings.LastIndex(e.Caller.File, "/")
	e.Caller.File = e.Caller.File[idx+1:]
	return nil
}

func InitLogrus(level logrus.Level) {
	logrus.SetLevel(level)
	logrus.SetReportCaller(true)
	logrus.AddHook(&DefaultFieldHook{})
}
