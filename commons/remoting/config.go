package remoting

import (
	"encoding/json"
)

type RemotingConfig struct {
	//Asynchronously send message size
	SendLimit int `json:"sendLimit" yaml:"sendLimit"`

	// the limit of packet send channel
	PacketBytesLimit int `json:"packetBytesLimit" yaml:"packetBytesLimit"`

	AcceptTimeout int `json:"acceptTimeout" yaml:"acceptTimeout"`

	//heartbeat time,and timeout SECONDS
	IdleTime int `json:"idleTime" yaml:"idleTime"`

	//连续几次heartbeat不传递就就认为掉线
	IdleTimeout int `json:"idleTimeout" yaml:"idleTimeout"`
}

func (cfg *RemotingConfig) String() string {
	bs, _ := json.Marshal(cfg)
	return string(bs)
}

func DefaultConfig() *RemotingConfig {
	return &RemotingConfig{
		SendLimit:        10000,
		PacketBytesLimit: 1024,
		AcceptTimeout:    3,
		IdleTime:         15,
		IdleTimeout:      3,
	}
}
