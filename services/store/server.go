package store

import (
	"fmt"
	"github.com/ihaiker/tenured-go-server/commons"
	"github.com/ihaiker/tenured-go-server/commons/protocol"
	"github.com/ihaiker/tenured-go-server/commons/registry"
	_ "github.com/ihaiker/tenured-go-server/commons/registry/consul"
	"github.com/ihaiker/tenured-go-server/services"
	"github.com/kataras/iris/core/errors"
)

type storeServer struct {
	config   *storeConfig
	address  string
	server   *protocol.TenuredServer
	registry registry.ServiceRegistry

	serviceWappers *ServicesWapper
}

func (this *storeServer) startTenuredServer() (err error) {
	if this.address, err = this.config.Tcp.GetAddress(); err != nil {
		return err
	}

	if this.server, err = protocol.NewTenuredServer(this.address, this.config.Tcp.RemotingConfig); err != nil {
		return err
	}

	this.server.AuthHeader = &protocol.AuthHeader{
		Module:     fmt.Sprintf("%s_%s", this.config.Prefix, "store"),
		Address:    this.address,
		Attributes: this.config.Tcp.Attributes,
	}

	this.serviceWappers.SetTCPServer(this.server)
	if err = this.serviceWappers.Start(); err != nil {
		return
	}

	return this.server.Start()
}

func (this *storeServer) startRegistry() error {
	//获取注册中心
	pluginsConfig, err := registry.ParseConfig(this.config.Registry.Address)
	if err != nil {
		return err
	}

	plugins, has := registry.GetPlugins(pluginsConfig.Plugin)
	if !has {
		return errors.New("not found registry: " + this.config.Registry.Address)
	}

	if this.registry, err = plugins.Registry(*pluginsConfig); err != nil {
		return err
	}

	//获取集群ID
	clusterId := services.NewClusterID(this.config.WorkDir, this.registry)

	if serverInstance, err := plugins.Instance(this.config.Registry.Attributes); err != nil {
		return err
	} else {
		serverInstance.Name = this.config.Prefix + "_store"
		serverInstance.Id, err = clusterId.Id(serverInstance.Name)
		if err != nil {
			return err
		}
		serverInstance.Address = this.address
		serverInstance.Metadata = this.config.Registry.Metadata
		serverInstance.Tags = this.config.Registry.Tags

		if err := this.registry.Register(*serverInstance); err != nil {
			return err
		}

		err = clusterId.CheckAndWrite(serverInstance.Name, serverInstance.Id)
	}
	return err
}

func (this *storeServer) Start() error {
	logger.Info("start store server.")
	if err := this.startTenuredServer(); err != nil {
		return err
	}
	if err := this.startRegistry(); err != nil {
		return err
	}
	return nil
}

func (this *storeServer) Shutdown(interrupt bool) {
	logger.Info("stop store server.")
	commons.ShutdownIfService(this.registry, interrupt)
	this.serviceWappers.Shutdown(interrupt)
	this.server.Shutdown(interrupt)
}

func newStoreServer(config *storeConfig) *storeServer {
	return &storeServer{config: config, serviceWappers: NewServicesWapper(config)}
}
