package services

import (
	"encoding/json"
	"errors"
	"github.com/go-yaml/yaml"
	"github.com/ihaiker/tenured-go-server/commons"
	_ "github.com/ihaiker/tenured-go-server/commons/logs"
	"github.com/ihaiker/tenured-go-server/commons/nets"
	"github.com/ihaiker/tenured-go-server/commons/remoting"
	"github.com/ihaiker/tenured-go-server/commons/runtime"
	"github.com/sirupsen/logrus"
	"path/filepath"
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

type Logs struct {
	Level   string `json:"level" yaml:"level"`
	Path    string `json:"path" yaml:"path"`
	Output  string `json:"Output" yaml:"Output"` //stdout,file
	Archive bool   `json:"archive" yaml:"archive"`
}

func SearchServerConfig(serverName string) []string {
	searchFiles := []string{
		runtime.GetWorkDir() + "/conf/" + serverName + ".yaml",
		runtime.GetWorkDir() + "/conf/" + serverName + ".json",
		runtime.GetWorkDir() + "/../conf/" + serverName + ".yaml",
		runtime.GetWorkDir() + "/../conf/" + serverName + ".json",
		"/etc/tenured/conf/" + serverName + ".yaml",
		"/etc/tenured/conf/" + serverName + ".json",
	}
	for k, v := range searchFiles {
		searchFiles[k], _ = filepath.Abs(v)
	}
	return searchFiles
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

	ext := filepath.Ext(path)
	switch ext {
	case ".json":
		if err := json.Unmarshal(bs, config); err != nil {
			return err
		}
	case ".yaml":
		if err := yaml.Unmarshal(bs, config); err != nil {
			return err
		}
	default:
		return errors.New("not support config file: " + path)
	}
	return nil
}
func LoadServerConfig(server, configFile string, configObj interface{}) error {
	if configFile != "" {
		return LoadConfig(configFile, configObj)
	} else {
		searchConfigs := SearchServerConfig(server)
		for _, searchConfig := range searchConfigs {
			if err := LoadConfig(searchConfig, configObj); err == nil {
				logrus.Info("use config file: ", searchConfig)
				return nil
			} else {
				logrus.Debugf("config file %s error: %s", searchConfig, err)
			}
		}
		return errors.New("any config found ! \n\t" + strings.Join(searchConfigs, "\n\t"))
	}
}
