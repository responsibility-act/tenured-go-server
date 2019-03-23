package command

type EnableStatus string

const APPLY = EnableStatus("apply")     //申请状态
const OK = EnableStatus("ok")           //OK正常
const DENY = EnableStatus("deny")       //拒绝
const Disable = EnableStatus("disable") //禁用

type Account struct {
	ID          string //申请账户的ID
	Name        string //企业名称，对接的企业名称
	Description string //企业描述

	Password string //账户密码

	BusinessLicense string `json:"businessLicense"` //企业营业执照
	Phone           string //企业联系人手机
	Email           string //企业联系人邮箱

	EnableStatus      EnableStatus `json:"enableStatus"`      //审核状态
	EnableDescription string       `json:"enableDescription"` //审核结果描述
	EnableTime        string       `json:"enableTime"`        //审核时间

	CreateTime string `json:"createTime"` //企业创建时间
}
