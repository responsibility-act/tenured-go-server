package tenant

import (
	"github.com/ihaiker/tenured-go-server/commons/mixins"
	"github.com/ihaiker/tenured-go-server/commons/nets"
	"github.com/ihaiker/tenured-go-server/engine"
	"github.com/ihaiker/tenured-go-server/services"
)

type TenantConfig struct {
	HTTP *nets.IpAndPort `json:"http"` //http监听地址

	Prefix string `json:"prefix" yaml:"prefix"` //注册服务的前缀，所有系统保持一致

	Data string `json:"data" yaml:"data"` //数据存储位置

	Logs *services.Logs `json:"logs" json:"logs"`

	Registry *services.Registry `json:"registry" yaml:"registry"` //注册中心

	StoreClient *engine.StoreEngineConfig `json:"storeClient" yaml:"storeClient"`
}

func NewTenantConfig() *TenantConfig {
	return &TenantConfig{
		HTTP: &nets.IpAndPort{
			Port:           mixins.PortTenant,
			EnableAutoPort: true,
		},
		Prefix: mixins.Get(mixins.KeyServerPrefix, mixins.ServerPrefix),
		Data:   mixins.Get(mixins.KeyDataPath, mixins.DataPath),
		Logs: &services.Logs{
			Level:  "info",
			Path:   mixins.Get(mixins.KeyDataPath, mixins.DataPath) + "/logs/tenant.log",
			Output: "stdout",
		},
		Registry: &services.Registry{
			Address: mixins.Get(mixins.KeyRegistry, mixins.Registry),
			Attributes: map[string]string{
				"checkType": "http",
				"health":    "/health",
			},
		},
		StoreClient: &engine.StoreEngineConfig{
			Type: "leveldb",
		},
	}
}
