package registry

type LoadBalance interface {
	//注册服务的列表，
	Select(obj ...interface{}) (serverInstances []*ServerInstance, regKey string, err error)
	//返回
	Return(regKey string)
}
