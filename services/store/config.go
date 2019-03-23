package store

import (
	"encoding/json"
	"errors"
	"github.com/ihaiker/tenured-go-server/commons"
	"github.com/ihaiker/tenured-go-server/commons/mixins"
	"github.com/ihaiker/tenured-go-server/commons/nets"
	"github.com/ihaiker/tenured-go-server/commons/remoting"
)

type storeConfig struct {
	Registry string `json:"registry" yaml:"registry"` //注册中心
	Prefix   string `json:"prefix" yaml:"prefix"`     //注册服务的前缀，所有系统保持一致

	Data string `json:"data" yaml:"data"` //数据存储位置

	Attributes map[string]string `json:"attributes" yaml:"attributes"`

	*nets.IpAndPort
	*remoting.RemotingConfig

	//服务附加属性
	Metadata map[string]string `json:"metadata" yaml:"metadata"`
	Tags     []string          `json:"tags" yaml:"tags"`
}

func initConfig(configPath string) (*storeConfig, error) {
	storeCfg = &storeConfig{
		Registry: mixins.Get(mixins.KeyRegistry, mixins.Registry),
		Prefix:   mixins.Get(mixins.KeyServerPrefix, mixins.ServerPrefix),
		IpAndPort: &nets.IpAndPort{
			Port: 6072,
		},
		RemotingConfig: remoting.DefaultConfig(),
	}

	if fs := commons.NewFile(configPath); !fs.Exist() || fs.IsDir() {
		return nil, errors.New("the config not found : " + configPath)
	} else if bs, err := fs.ToBytes(); err != nil {
		return nil, err
	} else if err := json.Unmarshal(bs, storeCfg); err != nil {
		return nil, err
	} else {
		return storeCfg, nil
	}
}
