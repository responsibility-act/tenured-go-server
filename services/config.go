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
	"os"
	"path/filepath"
	"regexp"
	"strconv"
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

type ExecutorParam struct {
	Type  string
	Param []int
}

type Executors map[string]string

func (this *Executors) Get(key string) (*ExecutorParam, bool) {
	if val, has := (*this)[key]; has {
		m := regexp.MustCompile(`(fix|single|scheduled)\((\d+),?(\d+)?\)`)
		gs := m.FindStringSubmatch(val)

		param := make([]int, len(gs[2:]))
		for i := 0; i < len(gs[2:]); i++ {
			param[i], _ = strconv.Atoi(gs[2+i])
		}
		return &ExecutorParam{Type: gs[1], Param: param}, true
	}
	return nil, false
}

type Logs struct {
	Level   string            `json:"level" yaml:"level"`
	Path    string            `json:"path" yaml:"path"`
	Output  string            `json:"output" yaml:"output"` //stdout,file
	Archive bool              `json:"archive" yaml:"archive"`
	Loggers map[string]string `json:"loggers" yaml:"loggers"`
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
		return os.ErrNotExist
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
		logrus.Debug("search config file.")
		searchConfigs := SearchServerConfig(server)
		for _, searchConfig := range searchConfigs {
			if err := LoadConfig(searchConfig, configObj); err == nil {
				logrus.Debug("use config file: ", searchConfig)
				return nil
			} else if err == os.ErrNotExist {
				logrus.Debug("the file not found: ", searchConfig)
			} else {
				return err
			}
		}
		logrus.Info("use default config.")
		return nil
	}
}
