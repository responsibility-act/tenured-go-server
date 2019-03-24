package store

import (
	"fmt"
	"github.com/ihaiker/tenured-go-server/api"
	"github.com/ihaiker/tenured-go-server/commons"
	"github.com/ihaiker/tenured-go-server/commons/protocol"
	"github.com/ihaiker/tenured-go-server/commons/registry"
	_ "github.com/ihaiker/tenured-go-server/commons/registry/consul"
	"github.com/kataras/iris/core/errors"
	"github.com/sirupsen/logrus"
)

type storeServer struct {
	config        *storeConfig
	address       string
	server        *protocol.TenuredServer
	registry      registry.ServiceRegistry
	accountServer api.AccountService
}

func (this *storeServer) initTenuredServer() (err error) {
	if this.address, err = this.config.GetAddress(); err != nil {
		return err
	}

	if this.server, err = protocol.NewTenuredServer(this.address, this.config.RemotingConfig); err != nil {
		return err
	}

	this.server.AuthHeader = &protocol.AuthHeader{
		Module:     fmt.Sprintf("%s_%s", this.config.Prefix, "store"),
		Address:    this.address,
		Attributes: this.config.Attributes,
	}

	if this.accountServer, err = handlerAccountServer(this.server); err != nil {
		return err
	}
	if ss, ok := this.accountServer.(commons.Service); ok {
		if err = ss.Start(); err != nil {
			return err
		}
	}
	return this.server.Start()
}

func (this *storeServer) initRegistry() error {
	pluginsConfig, err := registry.ParseConfig(this.config.Registry)
	if err != nil {
		return err
	}

	plugins, has := registry.GetPlugins(pluginsConfig.Plugin)
	if !has {
		return errors.New("not found registry: " + this.config.Registry)
	}

	if this.registry, err = plugins.Registry(*pluginsConfig); err != nil {
		return err
	}

	if ss, ok := this.registry.(commons.Service); ok {
		if err := ss.Start(); err != nil {
			return err
		}
	}

	if serverInstance, err := plugins.Instance(this.config.RegistryAttributes); err != nil {
		return err
	} else {
		serverInstance.Name = this.config.Prefix + "_store"
		serverInstance.Id = "4787dc7f-6a0f-41a6-92d6-d0c15e4a4c30" //TODO 这里的管理器需要修改
		serverInstance.Address = this.address
		serverInstance.Metadata = this.config.Metadata
		serverInstance.Tags = this.config.Tags

		if err := this.registry.Register(*serverInstance); err != nil {
			return err
		}
	}
	return nil
}

func (this *storeServer) Start() error {
	logrus.Info("start store server.")
	if err := this.initTenuredServer(); err != nil {
		return err
	}
	if err := this.initRegistry(); err != nil {
		return err
	}
	return nil
}

func (this *storeServer) Shutdown(interrupt bool) {
	logrus.Info("stop store server.")
	if ss, ok := this.accountServer.(commons.Service); ok {
		ss.Shutdown(false)
	}
	if ss, ok := this.registry.(commons.Service); ok {
		ss.Shutdown(false)
	}
	this.server.Shutdown(false)
}

func newStoreServer(config *storeConfig) *storeServer {
	return &storeServer{config: config}
}
