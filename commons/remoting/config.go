package remoting

import (
	"encoding/json"
)

type RemotingConfig struct {
	Listen string
	// the limit of packet send channel
	PacketBytesLimit int `json:"packed_bytes_limit" yaml:"packed_bytes_limit"`

	AcceptTimeout int `json:"accept_timeout" yaml:"accept_timeout"`

	//heartbeat time,and timeout
	IdleTime int `json:"idle_time" yaml:"idle_time"`

	//连续几次heartbeat不传递就就认为掉线
	IdleTimeout int `json:"idle_timeout" yaml:"idle_timeout"`
}

func (cfg *RemotingConfig) String() string {
	bs, _ := json.Marshal(cfg)
	return string(bs)
}

func DefaultConfig() *RemotingConfig {
	return &RemotingConfig{
		Listen:           ":6071",
		PacketBytesLimit: 1024,
		AcceptTimeout:    3,
		IdleTime:         15 * 1000,
		IdleTimeout:      3,
	}
}
