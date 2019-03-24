package account

import (
	"github.com/ihaiker/tenured-go-server/api"
	"github.com/ihaiker/tenured-go-server/api/command"
	"github.com/ihaiker/tenured-go-server/commons"
	"github.com/ihaiker/tenured-go-server/commons/registry"
	_ "github.com/ihaiker/tenured-go-server/commons/registry/consul"
	"github.com/kataras/iris/core/errors"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var reg registry.ServiceRegistry
var server api.AccountService

func Init() error {
	if config, err := registry.ParseConfig("consul://127.0.0.1:8500"); err != nil {
		return err
	} else if plugins, has := registry.GetPlugins(config.Plugin); !has {
		return errors.New("no registry")
	} else {
		if reg, err = plugins.Registry(*config); err != nil {
			return err
		}
	}
	server, _ = NewAccountServer("tenured_store", reg)
	return nil
}

func TestNewAccount(t *testing.T) {
	err := Init()
	assert.Nil(t, err)
	account := &command.Account{}
	account.Email = "wo@renzhen.la"
	start := time.Now()
	for i := 0; i < 1000; i++ {
		if acc, err := server.Apply(account); err != nil {
			t.Log(acc, err)
		}
	}
	t.Log(time.Now().UnixNano() - start.UnixNano())
	server.(commons.Service).Shutdown(true)
}
