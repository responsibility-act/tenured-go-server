package logs

import (
	"errors"
	"github.com/sirupsen/logrus"
)

var loggers map[string]*logrus.Logger
var debug bool

func init() {
	loggers = map[string]*logrus.Logger{}
	_ = InitLogger(map[string]string{}, "info", "stdout", "", false)
}

func GetLoggers() map[string]*logrus.Logger {
	return loggers
}

func GetLogger(loggerName string) *logrus.Logger {
	if loggerName == "" {
		loggerName = "root"
	}
	logger, has := loggers[loggerName]
	if has {
		return logger
	}
	rootLogger := loggers["root"]
	logger = logrus.New()

	logger.SetReportCaller(true)
	logger.SetFormatter(rootLogger.Formatter)
	logger.SetOutput(rootLogger.Out)
	logger.SetLevel(rootLogger.Level)
	loggers[loggerName] = logger
	return logger
}

func SetLogger(loggerName, level string) error {
	logrusLevel, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}
	if loggerName == "" {
		loggerName = "root"
	}
	if logger, has := loggers[loggerName]; has {
		logger.SetLevel(logrusLevel)
		return nil
	} else {
		return errors.New("not found logger!")
	}
}

func DebugLogger() {
	debug = true
	logrus.SetLevel(logrus.DebugLevel)
	for loggerName, logger := range GetLoggers() {
		logger.SetLevel(logrus.DebugLevel)
		logrus.Debugf("set logger:%s level: debug", loggerName)
	}
}

/*
初始化logs

loggerLevels定义logger及其level

level默认日志级别

output日志输出类型：stdout,file,socket

file 文件日志输出位置

archive 文件日志是否压缩

debug 是否是全部日志级别设置为debug
*/
func InitLogger(loggerLevels map[string]string, level, output, file string, archive bool) (err error) {
	outputWrite := logrus.StandardLogger().Out
	formatter := &TextFormatter{}

	if output == "file" {
		if outputWrite, err = NewRollingFileOutput(file, archive); err != nil {
			return err
		}
	}
	if loggerLevels == nil {
		loggerLevels = map[string]string{}
	}
	loggerLevels["root"] = level

	for loggerName, levelStr := range loggerLevels {
		level, err := logrus.ParseLevel(levelStr)
		if err != nil {
			return err
		}
		logger := loggers[loggerName]
		if logger == nil {
			if loggerName == "root" || loggerName == "" {
				logger = logrus.StandardLogger()
			} else {
				logger = logrus.New()
			}
		}

		if debug {
			logger.SetLevel(logrus.DebugLevel)
		} else {
			logger.SetLevel(level)
		}
		logger.SetReportCaller(true)
		logger.SetFormatter(formatter)
		logger.SetOutput(outputWrite)
		loggers[loggerName] = logger
	}
	return nil
}
