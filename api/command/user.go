package command

/*
	restful api account信息
*/

type UserType int

const Normal UserType = 0
const Kefu UserType = 1

type User struct {
	AppID   string `json:"appID"`   //应用ID
	CloudID string `json:"cloudID"` //云用户ID

	UserId   string            `json:"userID"`   //第三方系统ID，不允许修改、不允许超过32位
	NickName string            `json:"nickName"` //用户昵称，可以用于搜索
	Face     string            `json:"face"`     //用户头像
	Attrs    map[string]string `json:"attrs"`
	Type     UserType          `json:"type"` //账号类型，0:普通账号，1：客服账号
}
