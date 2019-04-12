package nets

import (
	"fmt"
)

type IpAndPort struct {
	//绑定地址
	Bind string `json:"bind" yaml:"bind"`

	//注册的外部地址
	External string `json:"external" yaml:"external"`

	//使用端口
	Port int `json:"port" yaml:"port"`
	//当端口被占用的时候是否可以自动寻找新的端口，开始位置是Port。
	EnableAutoPort bool `json:"enableAutoPort" yaml:"enableAutoPort"`
	autoPort       bool `json:"-"` //是否已经自动选择过了

	//忽略网络
	IgnoredInterfaces []string `json:"ignoredInterfaces" yaml:"ignoredInterfaces"`

	//倾向地址
	PreferredNetworks []string `json:"preferredNetworks" yaml:"preferredNetworks"`
}

func (this *IpAndPort) getPort(host string) (int, error) {
	var err error
	if this.EnableAutoPort {
		if this.autoPort {
			return this.Port, nil
		}
		this.Port, err = RandPort(host, this.Port, 65535)
	}
	return this.Port, err
}

func (this *IpAndPort) GetAddress() (string, error) {
	var err error = nil
	host := this.Bind
	if host == "" {
		host, err = GetLocalIP(this.IgnoredInterfaces, this.PreferredNetworks)
		if err != nil {
			return "", err
		}
	}
	port, err := this.getPort(host)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s:%d", host, port), nil
}

//获取公网地址
func (this *IpAndPort) GetExternal() (string, error) {
	var err error = nil
	host := this.External
	if host == "" {
		host, err = GetExternal()
		if err != nil {
			return "", err
		}
	}

	port, err := this.getPort(host)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s:%d", host, port), nil
}
