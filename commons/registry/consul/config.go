package consul

import (
	"github.com/ihaiker/tenured-go-server/commons/registry"
	"time"
)

type ConsulConfig struct {
	config registry.PluginConfig
}

func (this *ConsulConfig) Scheme() string {
	return this.config.Get("scheme", "http")
}

func (this *ConsulConfig) Address() string {
	return this.config.Address[0]
}

func (this *ConsulConfig) Datacenter() string {
	return this.config.Get("datacenter", "dc1")
}

func (this *ConsulConfig) Token() string {
	return this.config.Get("token", "")
}

func (this *ConsulConfig) HealthWaitTime() time.Duration {
	return time.Second * time.Duration(this.config.GetInt("healthWaitTime", 5))
}

func (this *ConsulConfig) HealthFailWaitTime() time.Duration {
	return time.Second * time.Duration(this.config.GetInt("failHealthWaitTime", 1))
}
