package command

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
