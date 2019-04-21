package consul

import (
	"github.com/ihaiker/tenured-go-server/commons"
	"github.com/ihaiker/tenured-go-server/commons/logs"
	"github.com/ihaiker/tenured-go-server/commons/registry"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var config *registry.PluginConfig

func init() {
	config, _ = registry.ParseConfig("consul://127.0.0.1:8500")
}

func TestConsulServiceRegistry(t *testing.T) {
	logs.DebugLogger()

	plugin, _ := NewRegistryPlugins(config)

	sr, err := plugin.Registry()
	assert.Nil(t, err)

	err = sr.Subscribe("test", func(serverInstances []*registry.ServerInstance) {
		logrus.Info("OnNotify deregister: ", serverInstances)
	})
	t.Log(err)

	si, _ := plugin.Instance(map[string]string{"interval": "1s"})

	si.Id = "b102c658-830a-4d63-ba08-6a1ab75823d8"
	si.Name = "test"
	si.Address = "127.0.0.1:6071"
	si.Metadata = map[string]string{"test_metadata": "demo"}

	err = sr.Register(si)
	assert.Nil(t, err)

	ss, err := sr.Lookup("test", nil)
	assert.Nil(t, err)
	for _, s := range ss {
		t.Log(s)
	}
	time.Sleep(time.Second * 5)

	err = sr.Unregister(si.Id)

	time.Sleep(time.Second * 5)
	(sr.(commons.Service)).Shutdown(true)

	assert.Nil(t, err)
}

func TestConsulServiceRegistry_Register(t *testing.T) {
	logs.DebugLogger()
	plugin, _ := NewRegistryPlugins(config)
	sr, err := plugin.Registry()
	assert.Nil(t, err)
	err = sr.Subscribe("tenured_store", func(serverInstances []*registry.ServerInstance) {
		logrus.Info("OnNotify : ", serverInstances)
	})
	t.Log(err)

	time.Sleep(time.Hour)
	(sr.(commons.Service)).Shutdown(true)
}
