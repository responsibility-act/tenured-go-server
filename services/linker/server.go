package linker

import (
	"fmt"
	"github.com/ihaiker/tenured-go-server/api/invoke"
	"github.com/ihaiker/tenured-go-server/commons"
	"github.com/ihaiker/tenured-go-server/commons/executors"
	"github.com/ihaiker/tenured-go-server/commons/mixins"
	"github.com/ihaiker/tenured-go-server/engine"
	"github.com/ihaiker/tenured-go-server/protocol"
	"github.com/ihaiker/tenured-go-server/registry"
	"github.com/ihaiker/tenured-go-server/registry/cache"
	"github.com/ihaiker/tenured-go-server/registry/load_balance"
	"github.com/ihaiker/tenured-go-server/registry/plugins"
	"hash/crc64"
)

type LinkerServer struct {
	address string

	config          *linkerConfig
	reg             registry.ServiceRegistry
	server          *protocol.TenuredServer
	serviceManager  commons.ServiceManager
	executorManager executors.ExecutorManager

	registryPlugin    registry.Plugins
	storeClientPlugin engine.StoreClientPlugin
	clientLoadBalance load_balance.LoadBalance
}

func NewLinkerServer(config *linkerConfig) *LinkerServer {
	return &LinkerServer{config: config}
}

func (this *LinkerServer) initStoreClientPlugin() (err error) {
	storeServerName := this.config.Prefix + "_store"
	if this.storeClientPlugin, err = engine.GetStoreClientPlugin(storeServerName, this.config.Engine, this.reg); err != nil {
		return err
	}
	this.serviceManager.Add(storeServerName)

	this.clientLoadBalance = this.storeClientPlugin.LoadBalance()
	this.serviceManager.Add(this.clientLoadBalance)
	return nil
}

func (this *LinkerServer) initTenuredServer() (err error) {
	if this.address, err = this.config.Tcp.GetAddress(); err != nil {
		return err
	}
	if this.server, err = protocol.NewTenuredServer(this.address, this.config.Tcp.RemotingConfig); err != nil {
		return err
	}
	this.server.AuthHeader = &protocol.AuthHeader{
		Module:  mixins.Linker(this.config.Prefix),
		Address: this.address,
	}
	if this.server.AuthChecker, err = NewLinkerAuthChecker(this.address, this.clientLoadBalance); err != nil {
		return err
	}
	this.server.SetSessionManager(protocol.NewMapSessionManager())
	this.serviceManager.Add(this.server)
	return nil
}

func (this *LinkerServer) initExecutorManager() error {
	this.executorManager = executors.NewExecutorManager(executors.NewFixedExecutorService(256, 10000))
	if err := this.executorManager.Config(this.config.Executors); err != nil {
		return err
	}
	this.serviceManager.Add(this.executorManager)
	return nil
}

func (this *LinkerServer) initRegistry() error {
	registryPlugins, err := plugins.GetRegistryPlugins(this.config.Registry.Address)
	if err != nil {
		return err
	}
	if reg, err := registryPlugins.Registry(); err != nil {
		return err
	} else {
		this.reg = cache.NewCacheRegistry(reg)
		this.serviceManager.Add(this.reg)
	}
	this.registryPlugin = registryPlugins
	this.serviceManager.Add(this.registryPlugin)
	return nil
}

func (this *LinkerServer) registryServer() error {
	serverName := this.config.Prefix + "_linker"
	if serverInstance, err := this.registryPlugin.Instance(this.config.Registry.Attributes); err != nil {
		return err
	} else {
		external, err := this.config.Tcp.GetExternal()
		if err != nil {
			return err
		}
		serverInstance.Name = serverName
		serverInstance.Id = fmt.Sprintf("%v", crc64.Checksum([]byte(this.address), crc64.MakeTable(crc64.ECMA)))
		serverInstance.Address = this.address
		serverInstance.Metadata = map[string]string{
			"external": external,
		}
		if err := this.reg.Register(serverInstance); err != nil {
			return err
		}
		return nil
	}
}

func (this *LinkerServer) registryCommandHandler() error {
	executorManager := executors.NewExecutorManager(executors.NewSingleExecutorService(1))
	this.serviceManager.Add(executorManager)
	invokeServer := NewLinkerCommandHandler(this.server.GetSessionManager())
	return invoke.NewLinkerServiceInvoke(this.server, invokeServer, executorManager)
}

func (this *LinkerServer) Start() (err error) {
	logger.Info("start linker server")
	if err = this.initExecutorManager(); err != nil {
		return
	}
	if err = this.initRegistry(); err != nil {
		return
	}
	if err = this.initTenuredServer(); err != nil {
		return
	}
	if err = this.initStoreClientPlugin(); err != nil {
		return
	}
	if err = this.registryCommandHandler(); err != nil {
		return
	}
	if err = this.serviceManager.Start(); err != nil {
		return
	}
	if err = this.registryServer(); err != nil {
		return
	}
	return nil
}

func (this *LinkerServer) Shutdown(interrupt bool) {
	logger.Info("shutdown linker server")
	this.serviceManager.Shutdown(interrupt)
}
