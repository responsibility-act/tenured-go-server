package api

import (
	"github.com/ihaiker/tenured-go-server/commons/protocol"
)

type Status string

const APPLY = Status("apply")     //申请状态
const OK = Status("ok")           //OK正常
const DENY = Status("deny")       //拒绝
const Disable = Status("disable") //禁用

type Account struct {
	ID          string `json:"id"`          //申请账户的ID
	Name        string `json:"name"`        //企业名称，对接的企业名称
	Description string `json:"description"` //企业描述

	Password string `json:"password"` //账户密码

	BusinessLicense string `json:"businessLicense"` //企业营业执照
	Phone           string `json:"phone"`           //企业联系人手机
	Email           string `json:"email"`           //企业联系人邮箱

	Status            Status `json:"status"`            //审核状态
	StatusDescription string `json:"statusDescription"` //审核结果描述
	StatusTime        string `json:"statusTime"`        //审核时间

	AllowIP []string `json:"allowIp" json:"allowIp"` //允许调用的IP地址

	CreateTime string `json:"createTime"` //企业创建时间
}

type App struct {
	ID          string `json:"id"`
	AccessKey   string `json:"accessKey"`
	SecurityKey string `json:"securityKey"`

	Domain string `json:"domain"` //分配给引用的专有域名

	Name        string `json:"name"`    //应用名称
	Icon        string `json:"appIcon"` //应用图标
	Description string `json:"description"`

	CreateTime string `json:"createTime"` //创建时间
}

//平台账户信息API
type AccountService interface {
	//申请一个账户
	Apply(applyAccount *Account) (account *Account, err *protocol.TenuredError)
}
