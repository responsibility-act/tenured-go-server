package logs

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestInitLogrus(t *testing.T) {
	err := InitLogrus("stdout", "debug", "/", true)
	assert.Nil(t, err)

	logrus.WithField("name", "value").WithField("name2", "value2").Info("test")
	logrus.WithField("name", "value").WithField("name2", "value2").Error("test")
	logrus.WithField("agent", "test").Info("测试")
}

func TestFileLogrus(t *testing.T) {
	err := InitLogrus("file", "debug", "./name.log", true)
	assert.Nil(t, err)

	logrus.WithField("name", "value").WithField("name2", "value2").Info("test")
	logrus.WithField("name", "value").WithField("name2", "value2").Error("test")
	logrus.WithField("agent", "test").Info("测试")
	logrus.Error("test========")

	for {
		<-time.After(time.Second)
		logrus.Info(time.Now().Format("2006-01-02 15:04:05"))
	}
}
