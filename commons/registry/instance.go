package registry

import (
	"fmt"
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

func (this ServerInstance) String() string {
	return fmt.Sprintf("[%s] %s(%s) %s", this.Status, this.Name, this.Address, this.Id)
}

func LoadModel(obj interface{}, m map[string]string) {
	if m == nil || len(m) == 0 {
		return
	}
	defer func() {
		if e := recover(); e != nil {
			logrus.Debug(e)
		}
	}()

	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	attrs := map[string]string{}
	for i := 0; i < val.Type().NumField(); i++ {
		name := val.Type().Field(i).Name
		if attr, has := val.Type().Field(i).Tag.Lookup("attr"); has {
			name = attr
		} else if jsonKey, has := val.Type().Field(i).Tag.Lookup("json"); has {
			name = jsonKey
		} else if yamlKey, has := val.Type().Field(i).Tag.Lookup("yaml"); has {
			name = yamlKey
		}
		attrs[name] = val.Type().Field(i).Name
	}

	for k, v := range m {
		if fieldName, has := attrs[k]; has {
			if f := val.FieldByName(fieldName); f.IsValid() {
				if f.CanSet() {
					switch f.Type().Kind() {
					case reflect.Int:
						if i, e := strconv.ParseInt(v, 0, 0); e == nil {
							f.SetInt(i)
						}
					case reflect.Float64:
						if fl, e := strconv.ParseFloat(v, 0); e == nil {
							f.SetFloat(fl)
						}
					case reflect.String:
						f.SetString(v)
					}
				}
			}
		}
	}
}

func IsOK(instance ServerInstance) bool {
	return instance.Status == "OK"
}

func AllNotOK(instance ...ServerInstance) bool {
	for _, v := range instance {
		if IsOK(v) {
			return false
		}
	}
	return true
}
