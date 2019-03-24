package protocol

import (
	"github.com/ihaiker/tenured-go-server/commons"
	"github.com/ihaiker/tenured-go-server/commons/executors"
	"github.com/ihaiker/tenured-go-server/commons/remoting"
	"github.com/kataras/iris/core/errors"
	"github.com/sirupsen/logrus"
	"reflect"
	"time"
)

type InvokeMethod struct {
	//0: func(requestCommand) responseCommand
	//1: func(header) header,error
	//2: func(header,body) header,body,error
	module int
	server interface{}
	method reflect.Method
	in     reflect.Type
	out    reflect.Type
}

func (this *InvokeMethod) invokeError(channel remoting.RemotingChannel, request *TenuredCommand, err error) {
	logrus.Error("handler error: ", err)
	response := NewACK(request.id)
	response.RemotingError(ErrorHandler(err))
	if err := channel.Write(response, time.Second*3); err != nil {
		logrus.Errorf("channel %s write message error: %s", channel.RemoteAddr(), err)
	}
}

//0: func(requestCommand) responseCommand
func (this *InvokeMethod) invoke0() TenuredCommandProcesser {
	return func(channel remoting.RemotingChannel, request *TenuredCommand) {
		defer func() {
			if err := recover(); err != nil {
				this.invokeError(channel, request, commons.Catch(err))
			}
		}()

		values := this.method.Func.Call([]reflect.Value{reflect.ValueOf(this.server), reflect.ValueOf(request)})
		if err := channel.Write(values[0], time.Second*3); err != nil {
			logrus.Errorf("channel %s write message error: %s", channel.RemoteAddr(), err)
		}
	}
}

//1: func(header) header,error
func (this *InvokeMethod) invoke1() TenuredCommandProcesser {
	return func(channel remoting.RemotingChannel, request *TenuredCommand) {
		defer func() {
			if err := recover(); err != nil {
				this.invokeError(channel, request, commons.Catch(err))
			}
		}()

		response := NewACK(request.id)
		requestHeader := reflect.New(this.in.Elem()).Interface()
		if err := request.GetHeader(requestHeader); err != nil {
			response.RemotingError(ErrorInvalidHeader(err))
		} else {
			values := this.method.Func.Call([]reflect.Value{reflect.ValueOf(this.server), reflect.ValueOf(requestHeader)})
			outHeader := values[0].Interface()
			err := values[1].Interface().(*TenuredError)
			if err != nil {
				response.RemotingError(ErrorHandler(err))
			} else if outHeader != nil {
				if err := response.SetHeader(outHeader); err != nil {
					response.RemotingError(ErrorHandler(err))
				}
			}
		}
		if err := channel.Write(response, time.Second*3); err != nil {
			logrus.Errorf("channel %s write message error: %s", channel.RemoteAddr(), err)
		}
	}
}

//2: func(header,body) header,body,error
func (this *InvokeMethod) invoke2() TenuredCommandProcesser {
	return func(channel remoting.RemotingChannel, request *TenuredCommand) {
		defer func() {
			if err := recover(); err != nil {
				this.invokeError(channel, request, commons.Catch(err))
			}
		}()
		response := NewACK(request.id)
		requestHeader := reflect.New(this.in.Elem()).Interface()
		requestBody := request.Body
		if err := request.GetHeader(requestHeader); err != nil {
			response.RemotingError(ErrorInvalidHeader(err))
		} else {
			values := this.method.Func.Call([]reflect.Value{
				reflect.ValueOf(this.server), reflect.ValueOf(requestHeader), reflect.ValueOf(requestBody),
			})
			outHeader := values[0].Interface()
			err := values[2].Interface().(*TenuredError)
			if err != nil {
				response.RemotingError(ErrorHandler(err))
			} else {
				if outHeader != nil {
					if err := response.SetHeader(outHeader); err != nil {
						response.RemotingError(ErrorHandler(err))
					}
				}
				response.Body = values[1].Bytes()
			}
		}
		if err := channel.Write(response, time.Second*3); err != nil {
			logrus.Errorf("channel %s write message error: %s", channel.RemoteAddr(), err)
		}
	}
}

func (this *InvokeMethod) Invoke() TenuredCommandProcesser {
	switch this.module {
	case 0:
		return this.invoke0()
	case 1:
		return this.invoke1()
	default:
		return this.invoke2()
	}
}

type TenuredInvoke struct {
	server  interface{}
	tenured TenuredService
}

func (this *TenuredInvoke) errors(method reflect.Method) error {
	return errors.New("Invoild method: " + method.Name + " " + method.Func.String())
}

func (this *TenuredInvoke) isBytes(t reflect.Type) bool {
	return t == reflect.TypeOf([]byte{})
}

func (this *TenuredInvoke) isError(t reflect.Type) bool {
	return t == reflect.TypeOf((*TenuredError)(nil))
}

func (this *TenuredInvoke) isCommand(t reflect.Type) bool {
	return t == reflect.TypeOf((*TenuredCommand)(nil))
}

func (this *TenuredInvoke) Invoke(code uint16, methodName string, executor executors.ExecutorService) error {
	serverInterface := reflect.TypeOf(this.server)

	method, has := serverInterface.MethodByName(methodName)
	if !has {
		return errors.New("method not found: " + methodName)
	}

	invokeMethod := &InvokeMethod{server: this.server, method: method}

	switch method.Type.NumIn() {
	case 2:
		inType := method.Type.In(1)
		if this.isCommand(inType) {
			if method.Type.NumOut() != 1 || !this.isCommand(method.Type.Out(0)) {
				return this.errors(method)
			}
			invokeMethod.module = 0
		} else if method.Type.NumOut() == 2 &&
			method.Type.In(1).Kind() == reflect.Ptr && this.isError(method.Type.Out(1)) {
			invokeMethod.module = 1
		} else {
			return this.errors(method)
		}
		invokeMethod.in = inType
	case 3:
		if method.Type.NumOut() == 3 &&
			this.isBytes(method.Type.In(2)) &&
			this.isBytes(method.Type.Out(1)) && this.isError(method.Type.Out(2)) {
			invokeMethod.module = 2
			invokeMethod.in = method.Type.In(1)
			invokeMethod.out = method.Type.Out(0)
		} else {
			return this.errors(method)
		}
	default:
		return this.errors(method)
	}

	this.tenured.RegisterCommandProcesser(code, invokeMethod.Invoke(), executor)
	return nil
}

func NewInvoke(tenured TenuredService, server interface{}) *TenuredInvoke {
	return &TenuredInvoke{tenured: tenured, server: server}
}
