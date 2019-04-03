package api

/*
	客户端连接认证信息
*/
type AuthHeader struct {
	AppId   string `json:"appID"`   //应用ID
	AppAK   string `json:"appAK"`   //应用AK
	Token   string `json:"token"`   //登录需要的token
	CloudID string `json:"cloudID"` //用户ID
	Sign    string `json:"sign"`    //安全校验值
}

type NotifyMode int

/**
 * 普通通知栏消息模式，离线消息下发后，单击通知栏消息直接启动应用，不会给应用进行回调
 */
const NotifyModeNormal NotifyMode = 0

/**
 * 自定义消息模式，离线消息下发后，单击通知栏消息会给应用进行回调
 */
const NotifyModeCustom NotifyMode = 1

type OfflinePush struct {
	//开启离线推送
	Enabled bool `json:"enabled"`

	//通知类型
	NotifyMode `json:"notifyMode"`

	//通知内容标题
	Tile string `json:"tile,omitempty"`

	/**
	 * 设置当前消息在对方收到离线推送时候展示内容（可选，发送消息时设置）
	 */
	Descr string `json:"descr,omitempty"`

	//离线声音提示
	AndroidRemindSound string `json:"androidRemindSound,omitempty"`
	IOSRemindSound     string `json:"iosRemindSound,omitempty"`

	//设置当前消息是否开启 Badge 计数，默认开启（可选，发送消息时设置）
	BadgeEnabled bool `json:"badgeEnabled"`

	/**
	 * 设置当前消息的扩展字段（可选，发送消息的时候设置）
	 */
	Ext string `json:"ext,omitempty"`
}

type commonsHeader struct {
	To                string       `json:"to"`                          //消息接收方
	EnableOfflinePush bool         `json:"enableOfflinePush,omitempty"` //开启离线通知，如果设置OfflinePush将覆盖默认值
	OfflinePush       *OfflinePush `json:"offlinePush,omitempty"`       //离线推送设置
}

type PeerTextHeader struct {
	commonsHeader
	Text string `json:"text"` //消息内容
}

type PeerImageHeader struct {
	commonsHeader
	Image string `json:"image"` //图片地址
}
