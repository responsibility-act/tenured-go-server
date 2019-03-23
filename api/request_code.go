package api

/*
	点对点消息。
	此处定义不使用itoa，担心后续增加打乱顺序
*/

const (
	Text      uint16 = 1002 //点对点文本消息
	Image     uint16 = 1003 //点对点图片
	Emoji     uint16 = 1004 //点对点表情
	Voice     uint16 = 1005 //点对点语音
	Video     uint16 = 1006 //点对点小视频
	Share     uint16 = 1007 //点对点分享
	File      uint16 = 1008 //点对点文件
	Locaation uint16 = 1009 //位置信息

	ApplyFriend     uint16 = 1010 //好友申请
	AllowFriend     uint16 = 1011 //同意好友申请
	DenyFriend      uint16 = 1012 //拒绝好友申请
	BlacklistFriend uint16 = 1013 //加入黑名单
	DeleteFriend    uint16 = 1014 //删除好友，被删除好友
)

const (
	AccountServiceApply = uint16(3001)
)
