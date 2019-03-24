package command

type Status string

const APPLY = Status("apply")     //申请状态
const OK = Status("ok")           //OK正常
const DENY = Status("deny")       //拒绝
const Disable = Status("disable") //禁用

type Account struct {
	ID          string //申请账户的ID
	Name        string //企业名称，对接的企业名称
	Description string //企业描述

	Password string //账户密码

	BusinessLicense string `json:"businessLicense"` //企业营业执照
	Phone           string //企业联系人手机
	Email           string //企业联系人邮箱

	Status            Status `json:"status"`            //审核状态
	StatusDescription string `json:"statusDescription"` //审核结果描述
	StatusTime        string `json:"statusTime"`        //审核时间

	CreateTime string `json:"createTime"` //企业创建时间
}
