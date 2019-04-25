package linker

import (
	"fmt"
	"github.com/ihaiker/tenured-go-server/commons"
	"github.com/ihaiker/tenured-go-server/commons/executors"
	"github.com/ihaiker/tenured-go-server/protocol"
	"github.com/ihaiker/tenured-go-server/registry"
	"github.com/ihaiker/tenured-go-server/registry/cache"
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

	registryPlugin registry.Plugins
}

func NewLinkerServer(config *linkerConfig) *LinkerServer {
	return &LinkerServer{config: config}
}

func (this *LinkerServer) initTenuredServer() (err error) {
	if this.address, err = this.config.Tcp.GetAddress(); err != nil {
		return err
	}
	if this.server, err = protocol.NewTenuredServer(this.address, this.config.Tcp.RemotingConfig); err != nil {
		return err
	}
	this.server.AuthHeader = &protocol.AuthHeader{
		Module:  fmt.Sprintf("%s_%s", this.config.Prefix, "linker"),
		Address: this.address,
	}
	if this.server.AuthChecker, err = NewLinkerAuthChecker(); err != nil {
		return err
	}
	this.serviceManager.Add(this.server)
	return nil
}

func (this *LinkerServer) initExecutorManager() {
	//TODO 需要实现 scheduled queue
	this.executorManager = executors.NewExecutorManager(executors.NewFixedExecutorService(256, 10000))
	for k, _ := range this.config.Executors {
		if ek, has := this.config.Executors.Get(k); has {
			switch ek.Type {
			case "fix":
				this.executorManager.Fix(k, ek.Param[0], ek.Param[1])
			case "single":
				this.executorManager.Single(k, ek.Param[0])
			case "scheduled":

			}
		}
	}
	this.serviceManager.Add(this.executorManager)
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

func (this *LinkerServer) Start() (err error) {
	logger.Info("start linker server")
	this.initExecutorManager()
	if err = this.initRegistry(); err != nil {
		return
	}
	if err = this.initTenuredServer(); err != nil {
		return
	}
	if err = this.registryServer(); err != nil {
		return
	}
	return this.serviceManager.Start()
}

func (this *LinkerServer) Shutdown(interrupt bool) {
	logger.Info("shutdown linker server")
	this.serviceManager.Shutdown(interrupt)
}
