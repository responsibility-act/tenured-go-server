package logs

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInitLogrus(t *testing.T) {
	err := InitLogrus("stdout", "debug", "/", true)
	assert.Nil(t, err)

	logrus.WithField("name", "value").WithField("name2", "value2").Info("test")
	logrus.WithField("name", "value").WithField("name2", "value2").Error("test")
	logrus.WithField("agent", "test").Info("测试")
}
