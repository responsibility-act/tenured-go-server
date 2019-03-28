package logs

import (
	"github.com/sirupsen/logrus"
	"io"
)

var loggers map[string]*logrus.Logger
var fileout map[string]io.Writer

func init() {
	fileout = map[string]io.Writer{}
	loggers = map[string]*logrus.Logger{}
	_, _ = InitLogger("", "stdout", "info", "", false)
}

func GetLogger(module string) (*logrus.Logger, bool) {
	logger, has := loggers[module]
	if !has {
		if module == "" {
			logger = logrus.StandardLogger()
		} else {
			logger = logrus.New()
		}

		loggers[module] = logger
	}
	return logger, has
}

func InitLogger(module, output, level, file string, archive bool) (*logrus.Logger, error) {
	logger, _ := GetLogger(module)

	logrusLevel, err := logrus.ParseLevel(level)
	if err != nil {
		return nil, err
	}
	logger.SetLevel(logrusLevel)
	logger.SetReportCaller(true)

	formatter := &TextFormatter{}
	logger.SetFormatter(formatter)

	if output == "file" {
		if write, has := fileout[file]; has {
			logger.SetOutput(write)
		} else {
			fileOutput, err := NewRollingFileOutput(file, archive)
			if err != nil {
				return nil, err
			} else {
				logger.SetOutput(fileOutput)
			}
		}
	}
	return logger, nil
}
