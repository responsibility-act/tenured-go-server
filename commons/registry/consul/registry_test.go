package consul

import (
	"github.com/ihaiker/tenured-go-server/commons"
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

func TestConsulServiceRegistry_Register(t *testing.T) {
	plugin, has := registry.GetPlugins(config.Plugin)
	assert.True(t, has)

	sr, err := plugin.Registry(*config)
	assert.Nil(t, err)

	err = sr.Subscribe("test", func(status registry.RegistionStatus, serverInstances []*registry.ServerInstance) {
		if status == registry.UNREGISTER {
			logrus.Info("OnNotify deregister: ", serverInstances)
		} else {
			logrus.Info("OnNotify register  : ", serverInstances)
		}
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
