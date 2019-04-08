package commons

import (
	"reflect"
	"sync/atomic"
)

const (
	S_STATUS_INIT       ServerStatus = 0 //服务初始化过程中
	S_STATUS_STARTING   ServerStatus = 1 //服务正在启动中
	S_STATUS_SUSPEND    ServerStatus = 2 //服务暂停中
	S_STATUS_RESTARTING ServerStatus = 3 //服务正在重启中
	S_STATUS_UP         ServerStatus = 4 //服务已经正常运行
	S_STATUS_STOPING    ServerStatus = 5 //服务正在停止
	S_STATUS_DOWN       ServerStatus = 6 //服务已经停止
)

type ServiceStarter interface {
	Start() error
}

type ServiceShutdowner interface {
	Shutdown(interrupt bool)
}

type Service interface {
	ServiceStarter
	ServiceShutdowner
}

func StartIfService(obj interface{}) error {
	if obj == nil || reflect.ValueOf(obj).IsNil() {
		return nil
	}
	if service, match := obj.(Service); match {
		return service.Start()
	}
	return nil
}

func ShutdownIfService(obj interface{}, interrupt bool) {
	if obj == nil || reflect.ValueOf(obj).IsNil() {
		return
	}
	if service, match := obj.(Service); match {
		service.Shutdown(interrupt)
	}
}

type ServiceManager struct {
	services []Service
}

func (this *ServiceManager) Add(objs ...interface{}) {
	if objs == nil {
		return
	}
	for _, obj := range objs {
		if service, match := obj.(Service); match {
			this.services = append(this.services, service)
		}
	}
}

func (this *ServiceManager) Start() error {
	for _, v := range this.services {
		if err := v.Start(); err != nil {
			return err
		}
	}
	return nil
}

func (this *ServiceManager) Shutdown(interrupt bool) {
	for _, v := range this.services {
		v.Shutdown(interrupt)
	}
}

func NewServiceManager() *ServiceManager {
	return &ServiceManager{services: make([]Service, 0)}
}

type ServerStatus uint32

func (this *ServerStatus) i() uint32 {
	return atomic.LoadUint32((*uint32)(this))
}

func (this ServerStatus) String() string {
	switch this {
	case S_STATUS_INIT:
		return "init"
	case S_STATUS_STARTING:
		return "starting"
	case S_STATUS_SUSPEND:
		return "suspend"
	case S_STATUS_RESTARTING:
		return "restarting"
	case S_STATUS_UP:
		return "up"
	case S_STATUS_STOPING:
		return "stoping"
	case S_STATUS_DOWN:
		return "down"
	}
	return "nnknow"
}

func (this *ServerStatus) change(old, new ServerStatus) bool {
	return atomic.CompareAndSwapUint32((*uint32)(this), old.i(), new.i())
}

func (this *ServerStatus) Is(status ServerStatus) bool {
	return atomic.LoadUint32((*uint32)(this)) == uint32(status)
}

func (this *ServerStatus) IsInit() bool {
	return this.Is(S_STATUS_INIT)
}

func (this *ServerStatus) IsStarting() bool {
	return this.Is(S_STATUS_STARTING) || this.Is(S_STATUS_RESTARTING)
}

func (this *ServerStatus) IsSuspend() bool {
	return this.Is(S_STATUS_SUSPEND)
}

func (this *ServerStatus) IsUp() bool {
	return this.Is(S_STATUS_UP)
}

func (this *ServerStatus) IsStoping() bool {
	return this.Is(S_STATUS_STOPING)
}

func (this *ServerStatus) IsDown() bool {
	return this.Is(S_STATUS_DOWN)
}

func (this *ServerStatus) Start(startFn func()) bool {
	if !this.change(S_STATUS_INIT, S_STATUS_STARTING) {
		return false
	}
	if startFn != nil {
		startFn()
	}
	return this.change(S_STATUS_STARTING, S_STATUS_UP)
}

func (this *ServerStatus) Suspend(suspendFn func()) bool {
	if !this.change(S_STATUS_UP, S_STATUS_SUSPEND) {
		return false
	}
	if suspendFn != nil {
		suspendFn()
	}
	return true
}

func (this *ServerStatus) ReStart(startFn func()) bool {
	if !this.change(S_STATUS_SUSPEND, S_STATUS_RESTARTING) {
		return false
	}
	if startFn != nil {
		startFn()
	}
	return this.change(S_STATUS_RESTARTING, S_STATUS_UP)
}

func (this *ServerStatus) Shutdown(shutdownFn func()) bool {
	if this.Is(S_STATUS_STOPING) || this.Is(S_STATUS_DOWN) {
		return false
	}
	if this.change(S_STATUS_UP, S_STATUS_STOPING) {
		if shutdownFn != nil {
			shutdownFn()
		}
	}
	return this.change(S_STATUS_STOPING, S_STATUS_DOWN)
}
