package command

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
