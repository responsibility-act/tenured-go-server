package console

import (
	"github.com/ihaiker/tenured-go-server/commons/mixins"
	"github.com/ihaiker/tenured-go-server/commons/nets"
	"github.com/ihaiker/tenured-go-server/commons/remoting"
	"github.com/ihaiker/tenured-go-server/commons/runtime"
	"github.com/ihaiker/tenured-go-server/services"
)

type ConsoleConfig struct {
	HTTP string `json:"http"` //http监听地址

	Prefix string `json:"prefix" yaml:"prefix"` //注册服务的前缀，所有系统保持一致

	Data string `json:"data" yaml:"data"` //数据存储位置

	WorkDir string `json:"workDir" json:"workDir"`

	Logs *services.Logs `json:"logs" json:"logs"`

	Registry *services.Registry `json:"registry" yaml:"registry"` //注册中心

	Tcp *services.Tcp `json:"tcp" yaml:"tcp"`

	Executors services.Executors `json:"executors"`
}

func NewConsoleConfig() *ConsoleConfig {
	return &ConsoleConfig{
		HTTP:    ":6074",
		Prefix:  mixins.Get(mixins.KeyServerPrefix, mixins.ServerPrefix),
		Data:    mixins.Get(mixins.KeyDataPath, mixins.DataPath),
		WorkDir: runtime.GetWorkDir(),
		Logs: &services.Logs{
			Level:  "info",
			Path:   mixins.Get(mixins.KeyDataPath, mixins.DataPath) + "/logs/console.log",
			Output: "stdout",
		},
		Registry: &services.Registry{
			Address: mixins.Get(mixins.KeyRegistry, mixins.Registry),
			Attributes: map[string]string{
				"CheckType": "http",
			},
		},
		Tcp: &services.Tcp{
			IpAndPort: &nets.IpAndPort{
				Port: mixins.GetInt("tenured.console.port", 6073),
			},
			RemotingConfig: remoting.DefaultConfig(),
		},
		Executors: services.Executors(map[string]int{}),
	}
}
