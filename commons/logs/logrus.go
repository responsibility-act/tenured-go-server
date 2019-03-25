package logs

import (
	"github.com/sirupsen/logrus"
)

func InitLogrus(output, level, file string, archive bool) error {
	logrusLevel, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}
	logrus.SetLevel(logrusLevel)
	logrus.SetReportCaller(true)

	formatter := &TextFormatter{}
	logrus.SetFormatter(formatter)

	if output == "file" {
		fileOutput, err := NewRollingFileOutput(file, archive)
		if err != nil {
			return err
		} else {
			logrus.SetOutput(fileOutput)
		}
	}
	return nil
}
