package store

import (
	"fmt"
	"github.com/ihaiker/tenured-go-server/commons"
	"github.com/ihaiker/tenured-go-server/commons/protocol"
	"github.com/ihaiker/tenured-go-server/commons/registry"
	_ "github.com/ihaiker/tenured-go-server/commons/registry/consul"
	"github.com/kataras/iris/core/errors"
	uuid "github.com/satori/go.uuid"
	"strings"
	"time"
)

type storeServer struct {
	config   *storeConfig
	address  string
	server   *protocol.TenuredServer
	registry registry.ServiceRegistry

	serviceInvokeManager *ServicesInvoke
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

	this.serviceInvokeManager = NewServicesWapper(this.config, this.server)
	if err = this.serviceInvokeManager.Start(); err != nil {
		return
	}
	return
}

func NewClusterID(workDir string, serverName string) (string, string, error) {
	clusterIdFile := commons.NewFile(workDir + fmt.Sprintf("/%s.cid", serverName))

	if clusterIdFile.Exist() {
		if line, err := clusterIdFile.ToString(); err != nil {
			return "", "", err
		} else {
			sp := strings.SplitN(line, ",", 2)
			return sp[0], sp[1], nil
		}
	}
	u, _ := uuid.NewV4()
	clusterId := strings.ToUpper(strings.ReplaceAll(u.String(), "-", ""))
	firstStartTime := time.Now().Format("20060102150405")
	if out, err := clusterIdFile.GetWriter(false); err != nil {
		return "", "", err
	} else {
		defer out.Close()
		if _, err := out.Write([]byte(clusterId + "," + firstStartTime)); err != nil {
			return "", "", err
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

	if this.registry, err = plugins.Registry(*pluginsConfig); err != nil {
		return err
	}

	//注册服务名称
	serverName := this.config.Prefix + "_store"

	//获取集群ID
	clusterId, _, err := NewClusterID(this.config.Data, serverName)
	if err != nil {
		return err
	}
	if serverInstance, err := plugins.Instance(this.config.Registry.Attributes); err != nil {
		return err
	} else {
		serverInstance.Name = serverName
		serverInstance.Id = clusterId
		serverInstance.Address = this.address
		serverInstance.Metadata = this.config.Registry.Metadata
		if serverInstance.Metadata == nil {
			serverInstance.Metadata = map[string]string{}
		}
		serverInstance.Metadata["FirstStartTime"] = "834953373" //for TimeHashLoadBalance
		serverInstance.Tags = this.config.Registry.Tags
		if err := this.registry.Register(*serverInstance); err != nil {
			return err
		}
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
