package store

import (
	"fmt"
	"github.com/ihaiker/tenured-go-server/api"
	"github.com/ihaiker/tenured-go-server/commons"
	"github.com/ihaiker/tenured-go-server/commons/executors"
	"github.com/ihaiker/tenured-go-server/commons/protocol"
	"github.com/ihaiker/tenured-go-server/commons/registry"
	"github.com/ihaiker/tenured-go-server/commons/registry/cache"
	"github.com/ihaiker/tenured-go-server/commons/remoting"
	"github.com/ihaiker/tenured-go-server/commons/snowflake"
	"github.com/ihaiker/tenured-go-server/plugins"
	"strconv"
	"time"
)

type storeServer struct {
	config  *storeConfig
	address string

	server          *protocol.TenuredServer
	registry        registry.ServiceRegistry
	registryPlugins registry.Plugins

	serviceInvokeManager *ServicesInvokeManager

	executorManager executors.ExecutorManager
	snowflakeId     *snowflake.Snowflake

	serviceManger *commons.ServiceManager
}

func (this *storeServer) init() error {
	this.initExecutorManager()

	if err := this.initRegistry(); err != nil {
		return err
	}
	if err := this.initTenuredServer(); err != nil {
		return err
	}
	if err := this.initServicesInvoke(); err != nil {
		return err
	}
	return nil
}

func (this *storeServer) initExecutorManager() {
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
	this.serviceManger.Add(this.executorManager)
}

func (this *storeServer) initTenuredServer() (err error) {
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

	this.serviceManger.Add(this.server)
	return nil
}

func (this *storeServer) initServicesInvoke() (err error) {
	this.serviceInvokeManager = NewServicesInvokeManager(this.config, this.server, this.executorManager)
	this.serviceManger.Add(this.serviceInvokeManager)

	this.server.RegisterCommandProcesser(api.ClusterIdServiceGet, func(channel remoting.RemotingChannel, request *protocol.TenuredCommand) {
		logger.Debugf("Get clusterId: %s", channel.RemoteAddr())
		response := protocol.NewACK(request.ID())
		id, _ := this.snowflakeId.NextID()
		response.Body = commons.UInt64(id)
		if err := channel.Write(response, time.Millisecond*3000); err != nil {
			logger.Error("snowflake write error: ", err)
		}
	}, this.executorManager.Get("Snowflake"))

	return nil
}

func (this *storeServer) maxMachineId(serverName string) (uint16, error) {
	if ss, err := this.registry.Lookup(serverName, nil); err != nil {
		return 0, err
	} else {
		maxMachineId := uint16(0)
		for _, s := range ss {
			serId, _ := strconv.ParseUint(s.Id, 10, 64)
			p := snowflake.Decompose(serId)
			if p.MachineId > maxMachineId {
				maxMachineId = p.MachineId
			}
		}
		return maxMachineId, nil
	}
}

//return clusterId(snowflakeId)
func (this *storeServer) ClusterID(serverName string) (uint64, uint64, error) {
	clusterId := uint64(0)
	firstStartTime := uint64(0)

	clusterIdFile := commons.NewFile(this.config.Data + fmt.Sprintf("/%s.cid", serverName))
	if clusterIdFile.Exist() {
		if line, err := clusterIdFile.ToString(); err != nil {
			return 0, 0, err
		} else {
			clusterId, firstStartTime, _ = commons.SplitToUint2(line, 10, 64)
			machineId := snowflake.Decompose(clusterId).MachineId
			this.snowflakeId = snowflake.NewSnowflake(snowflake.Settings{
				MachineID: machineId,
			})
			return clusterId, firstStartTime, nil
		}
	}

	if maxMachineId, err := this.maxMachineId(serverName); err != nil {
		return 0, 0, err
	} else {
		this.snowflakeId = snowflake.NewSnowflake(snowflake.Settings{
			MachineID: maxMachineId + 1,
		})
	}

	if nextId, err := this.snowflakeId.NextID(); err != nil {
		return 0, 0, err
	} else {
		clusterId = nextId
		firstStartTime = snowflake.Decompose(clusterId).Time
	}

	if out, err := clusterIdFile.GetWriter(false); err != nil {
		return 0, 0, err
	} else {
		defer out.Close()
		if _, err := out.WriteString(fmt.Sprintf("%d,%d", clusterId, firstStartTime)); err != nil {
			return 0, 0, err
		}
	}
	return clusterId, firstStartTime, nil
}

func (this *storeServer) initRegistry() error {
	registryPlugins, err := plugins.GetRegistryPlugins(this.config.Registry.Address)
	if err != nil {
		return err
	}
	if reg, err := registryPlugins.Registry(); err != nil {
		return err
	} else {
		this.registry = cache.NewCacheRegistry(reg)
		this.serviceManger.Add(reg)
	}
	this.registryPlugins = registryPlugins
	return nil
}

func (this *storeServer) registryService() error {
	//注册服务名称
	serverName := this.config.Prefix + "_store"
	//获取集群ID
	clusterId, firstStartTime, err := this.ClusterID(serverName)
	if err != nil {
		return err
	}
	if serverInstance, err := this.registryPlugins.Instance(this.config.Registry.Attributes); err != nil {
		return err
	} else {
		serverInstance.Name = serverName
		serverInstance.Id = fmt.Sprintf("%d", clusterId)
		serverInstance.Address = this.address
		serverInstance.Metadata = this.config.Registry.Metadata
		if serverInstance.Metadata == nil {
			serverInstance.Metadata = map[string]string{}
		}
		serverInstance.Metadata["FirstStartTime"] = fmt.Sprintf("%d", firstStartTime)
		serverInstance.Tags = this.config.Registry.Tags
		if err := this.registry.Register(serverInstance); err != nil {
			return err
		}
	}
	return err
}

func (this *storeServer) Start() error {
	logger.Info("start store server.")
	if err := this.init(); err != nil {
		return err
	}
	if err := this.serviceManger.Start(); err != nil {
		return err
	}
	if err := this.registryService(); err != nil {
		return err
	}
	return nil
}

func (this *storeServer) Shutdown(interrupt bool) {
	logger.Info("stop store server.")
	this.serviceManger.Shutdown(interrupt)
}

func newStoreServer(config *storeConfig) *storeServer {
	return &storeServer{config: config, serviceManger: commons.NewServiceManager()}
}
