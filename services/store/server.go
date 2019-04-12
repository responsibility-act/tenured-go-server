package store

import (
	"fmt"
	"github.com/ihaiker/tenured-go-server/api"
	"github.com/ihaiker/tenured-go-server/commons"
	"github.com/ihaiker/tenured-go-server/commons/executors"
	"github.com/ihaiker/tenured-go-server/commons/protocol"
	"github.com/ihaiker/tenured-go-server/commons/registry"
	"github.com/ihaiker/tenured-go-server/commons/registry/cache"
	_ "github.com/ihaiker/tenured-go-server/commons/registry/consul"
	"github.com/ihaiker/tenured-go-server/commons/remoting"
	"github.com/ihaiker/tenured-go-server/commons/snowflake"
	"github.com/kataras/iris/core/errors"
	"strconv"
	"time"
)

type storeServer struct {
	config   *storeConfig
	address  string
	server   *protocol.TenuredServer
	registry registry.ServiceRegistry

	serviceInvokeManager *ServicesInvokeManager

	executorManager executors.ExecutorManager
	snowflakeId     *snowflake.Snowflake
}

func (this *storeServer) initExecutorManager() executors.ExecutorManager {
	manager := executors.NewExecutorManager(executors.NewFixedExecutorService(256, 10000))

	for k, _ := range this.config.Executors {
		if ek, has := this.config.Executors.Get(k); has {
			switch ek.Type {
			case "fix":
				manager.Fix(k, ek.Param[0], ek.Param[1])
			case "single":
				manager.Single(k, ek.Param[0])
			case "scheduled":

			}
		}
	}
	return manager
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

	if err = this.server.Start(); err != nil {
		return
	}

	this.executorManager = this.initExecutorManager()
	this.serviceInvokeManager = NewServicesInvokeManager(this.config, this.server, this.executorManager)
	if err = this.serviceInvokeManager.Start(); err != nil {
		return
	}
	return
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

	if reg, err := plugins.Registry(*pluginsConfig); err != nil {
		return err
	} else {
		this.registry = cache.NewCacheRegistry(reg)
	}

	//注册服务名称
	serverName := this.config.Prefix + "_store"
	//获取集群ID
	clusterId, firstStartTime, err := this.ClusterID(serverName)
	if err != nil {
		return err
	}
	if serverInstance, err := plugins.Instance(this.config.Registry.Attributes); err != nil {
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

func (this *storeServer) registrySnowflake() {
	this.server.RegisterCommandProcesser(api.ClusterIdServiceGet, func(channel remoting.RemotingChannel, request *protocol.TenuredCommand) {
		logger.Debugf("Get clusterId: %s", channel.RemoteAddr())
		response := protocol.NewACK(request.ID())
		id, _ := this.snowflakeId.NextID()
		response.Body = commons.UInt64(id)
		if err := channel.Write(response, time.Millisecond*3000); err != nil {
			logger.Error("snowflake write error: ", err)
		}
	}, this.executorManager.Get("Snowflake"))
}

func (this *storeServer) Start() error {
	logger.Info("start store server.")
	if err := this.startTenuredServer(); err != nil {
		return err
	}
	if err := this.startRegistry(); err != nil {
		return err
	}
	this.registrySnowflake()
	return nil
}

func (this *storeServer) Shutdown(interrupt bool) {
	logger.Info("stop store server.")
	commons.ShutdownIfService(this.registry, interrupt)
	if this.serviceInvokeManager != nil {
		this.serviceInvokeManager.Shutdown(interrupt)
	}
	if this.server != nil {
		this.server.Shutdown(interrupt)
	}
}

func newStoreServer(config *storeConfig) *storeServer {
	return &storeServer{config: config}
}
