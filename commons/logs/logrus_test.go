package logs

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInitLogrus(t *testing.T) {
	logger, err := InitLogger("test", "stdout", "debug", "/", true)
	assert.Nil(t, err)

	logger.WithField("name", "value").WithField("name2", "value2").Info("test")
	logger.WithField("name", "value").WithField("name2", "value2").Error("test")
	logger.WithField("agent", "test").Info("测试")
}

func TestFileLogrus(t *testing.T) {
	logger, err := InitLogger("file", "file", "debug", "./name.log", true)
	assert.Nil(t, err)

	logger.WithField("name", "value").WithField("name2", "value2").Info("test")
	logger.WithField("name", "value").WithField("name2", "value2").Error("test")
	logger.WithField("agent", "test").Info("测试")
	logger.Error("test========")
}
