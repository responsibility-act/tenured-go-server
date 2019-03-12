package registry

import (
	"github.com/sirupsen/logrus"
	"reflect"
	"strconv"
)

//注册中心服务实例附属属性，用于想不通的注册中心发送注册是附加参数的配置方式
type ServerInstanceAttrs interface {
	Config(map[string]string)
}

type ServerInstance struct {
	//serverID 全局唯一,使用uuid方式生成
	Id string

	//服务名称，例如：推送服务(push)，API服务（api）
	Name string

	//服务附加属性
	Metadata map[string]string

	//注册地址
	Address string

	PluginAttrs ServerInstanceAttrs

	Tags []string

	//当前状态,OK:正常状态，其他均为失败
	Status string
}

func LoadModel(obj interface{}, m map[string]string) {
	defer func() {
		if e := recover(); e != nil {
			logrus.Debug(e)
		}
	}()

	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	for k, v := range m {
		if f := val.FieldByName(k); f.IsValid() {
			if f.CanSet() {
				switch f.Type().Kind() {
				case reflect.Int:
					if i, e := strconv.ParseInt(v, 0, 0); e == nil {
						f.SetInt(i)
					} else {
						logrus.Debugf("Could not set int value of %s: %s\n", k, e)
					}
				case reflect.Float64:
					if fl, e := strconv.ParseFloat(v, 0); e == nil {
						f.SetFloat(fl)
					} else {
						logrus.Debugf("Could not set float64 value of %s: %s\n", k, e)
					}
				case reflect.String:
					f.SetString(v)

				default:
					logrus.Debugf("Unsupported format %v for field %s\n", f.Type().Kind(), k)
				}
			} else {
				logrus.Debugf("Key '%s' cannot be set\n", k)
			}
		} else {
			logrus.Debugf("Key '%s' does not have a corresponding field in obj %+v\n", k, obj)
		}
	}
}
