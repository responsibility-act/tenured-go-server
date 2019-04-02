package logs

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLogrus(t *testing.T) {
	logrus.Debug("not show")
	logrus.WithField("name", "value").WithField("name2", "value2").Info("test")
	logrus.WithField("name", "value").WithField("name2", "value2").Error("test")
	logrus.WithField("agent", "test").Info("测试")
	_ = SetLogger("", "debug")
	logrus.Debug("show .....")
}

func TestFileLogrus(t *testing.T) {
	err := InitLogger(map[string]string{}, "debug", "file", "./name.log", true, false)
	assert.Nil(t, err)

	logger := GetLogger("")

	logger.WithField("name", "value").WithField("name2", "value2").Info("test")
	logger.WithField("name", "value").WithField("name2", "value2").Error("test")
	logger.WithField("agent", "test").Info("测试")
	logger.Error("test========")
}
