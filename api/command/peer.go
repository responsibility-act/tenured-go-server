package command

type commonsHander struct {
	To                string       `json:"to"`                          //消息接收方
	EnableOfflinePush bool         `json:"enableOfflinePush,omitempty"` //开启离线通知，如果设置OfflinePush将覆盖默认值
	OfflinePush       *OfflinePush `json:"offlinePush,omitempty"`       //离线推送设置
}

type PeerTextHeader struct {
	commonsHander
	Text string `json:"text"` //消息内容
}

type PeerImageHeader struct {
	commonsHander
	Image string `json:"image"` //图片地址
}
