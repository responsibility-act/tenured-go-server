package registry

import (
	"net/url"
	"strconv"
	"strings"
)

//注册中心配置，此处配置类似于url的方式
type PluginConfig struct {
	Plugin  string
	Address []string
	Params  url.Values
	User    url.Userinfo
}

func (this *PluginConfig) GetInt(key string, def int) int {
	if value, ok := this.Params[key]; ok {
		if intValue, err := strconv.Atoi(value[0]); err == nil {
			return intValue
		}
	}
	return def
}

func (this *PluginConfig) Get(key, def string) string {
	if value, ok := this.Params[key]; ok {
		return value[0]
	} else {
		return def
	}
}

func (this *PluginConfig) Apply(key string, has func(value string)) {
	if value, ok := this.Params[key]; ok {
		has(value[0])
	}
}

func ParseConfig(cfg string) (*PluginConfig, error) {
	if u, err := url.Parse(cfg); err != nil {
		return nil, err
	} else {
		config := &PluginConfig{}
		config.Plugin = u.Scheme
		config.Address = strings.Split(u.Host, ";")
		if u.User != nil {
			config.User = *u.User
		}
		config.Params = u.Query()
		return config, nil
	}
}
