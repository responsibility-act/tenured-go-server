package remoting

import (
	"encoding/json"
)

type RemotingConfig struct {
	//Asynchronously send message size
	SendLimit int `json:"send_limit" yaml:"send_limit"`

	// the limit of packet send channel
	PacketBytesLimit int `json:"packed_bytes_limit" yaml:"packed_bytes_limit"`

	AcceptTimeout int `json:"accept_timeout" yaml:"accept_timeout"`

	//heartbeat time,and timeout SECONDS
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
		SendLimit:        1000,
		PacketBytesLimit: 1024,
		AcceptTimeout:    3,
		IdleTime:         15,
		IdleTimeout:      3,
	}
}
