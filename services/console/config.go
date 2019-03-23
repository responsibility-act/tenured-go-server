package console

type ConsoleConfig struct {
	HTTP string `json:"http"` //http监听地址
	TCP  string `json:"tcp"`  //tcp 监听地址

	DataCenter string `json:"dataCenter"` //数据存储位置
}
