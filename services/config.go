package services

import (
	"encoding/json"
	"errors"
	"github.com/go-yaml/yaml"
	"github.com/ihaiker/tenured-go-server/commons"
	"github.com/ihaiker/tenured-go-server/commons/nets"
	"github.com/ihaiker/tenured-go-server/commons/remoting"
	"github.com/ihaiker/tenured-go-server/commons/runtime"
	"strings"
)

type Registry struct {
	Address string `json:"address" yaml:"address"`
	//注册服务与注册中心的参数配置
	Attributes map[string]string `json:"attributes,omitempty" yaml:"attributes,omitempty"`

	//注册服务的元数据
	Metadata map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Tags     []string          `json:"tags,omitempty" yaml:"tags,omitempty"`
}

type Tcp struct {
	*nets.IpAndPort
	*remoting.RemotingConfig

	Attributes map[string]string `json:"attributes,omitempty" yaml:"attributes,omitempty"`
}

type Executors map[string]int

func (this *Executors) Get(key string, def int) int {
	if val, has := (*this)[key]; has {
		return val
	}
	return def
}

func SearchConfigs(serverName string) []string {
	return []string{
		runtime.GetWorkDir() + "/conf/" + serverName + ".yml",
		runtime.GetWorkDir() + "/conf/" + serverName + ".json",
		"/etc/tenured/conf/" + serverName + ".yml",
		"/etc/tenured/conf/" + serverName + ".json",
	}
}

func LoadConfig(path string, config interface{}) error {
	fs := commons.NewFile(path)
	if !fs.Exist() || fs.IsDir() {
		return errors.New("the config not found : " + path)
	}

	bs, err := fs.ToBytes()
	if err != nil {
		return err
	}

	if strings.HasSuffix(path, ".json") {
		if err := json.Unmarshal(bs, config); err != nil {
			return err
		}
	} else if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
		if err := yaml.Unmarshal(bs, config); err != nil {
			return err
		}
	} else {
		return errors.New("not support config file: " + path)
	}
	return nil
}
