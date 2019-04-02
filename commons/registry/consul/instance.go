package consul

import "github.com/ihaiker/tenured-go-server/commons/registry"

type ConsulServerAttrs struct {
	//检查类型
	CheckType string `json:"type" yaml:"type" attr:"type"` //http,tcp

	Health string `json:"health" yaml:"health" attr:"health"` //http url

	//检查频率，
	Interval string `json:"interval" yaml:"interval" attr:"interval"` //10s

	//当前节点出现异常多长时间下线
	Deregister string `json:"deregister" yaml:"deregister" attr:"deregister"`

	//请求处理超时时间
	RequestTimeout string `json:"requestTimeout" yaml:"requestTimeout" attr:"requestTime"`
}

func (this *ConsulServerAttrs) Config(attrs map[string]string) {
	registry.LoadModel(this, attrs)
}

func newInstance() *ConsulServerAttrs {
	return &ConsulServerAttrs{
		CheckType:      "tcp",
		Health:         "/health",
		Interval:       "5s",
		Deregister:     "15s",
		RequestTimeout: "3s",
	}
}
