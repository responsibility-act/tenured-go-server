package command

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
