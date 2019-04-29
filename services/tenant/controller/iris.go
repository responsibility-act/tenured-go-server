package ctl

import (
	"context"
	"github.com/ihaiker/tenured-go-server/api"
	"github.com/ihaiker/tenured-go-server/commons"
	"github.com/ihaiker/tenured-go-server/commons/logs"
	"github.com/ihaiker/tenured-go-server/protocol"
	"github.com/kataras/iris"
	ctx "github.com/kataras/iris/context"
	"time"
)

var app = iris.Default()
var logger = logs.GetLogger("ctrl")

var UserService api.UserService
var AccountService api.AccountService
var ClusterIdService api.ClusterIdService
var LinkerService api.LinkerService

type HttpServer struct {
	http           string
	serviceManager *commons.ServiceManager
}

func (this *HttpServer) Start() (err error) {
	logger.Debug("http server start: ", this.http)
	app.Logger().SetLevel(logger.Level.String())
	app.Logger().SetOutput(logger.Out)
	app.Logger().SetTimeFormat("2006-01-02 15:04:05")
	app.Logger().SetPrefix("(iris) ")

	app.Get("/health", func(ctx ctx.Context) {
		_, _ = ctx.JSON(map[string]interface{}{"status": "UP"})
	})

	startErr := make(chan error, 0)
	go func() {
		defer close(startErr)
		err = app.Run(
			iris.Addr(this.http),
			iris.WithoutBanner,
			iris.WithoutServerError(iris.ErrServerClosed),
		)
		startErr <- err
	}()

	select {
	case err = <-startErr:
		return
	case <-time.After(time.Second):
		return nil
	}
}

func (this *HttpServer) Shutdown(interrupt bool) {
	if err := app.Shutdown(context.Background()); err != nil {
		logger.Error("shutdown console http server error:", err)
	}
}

func NewHttpServer(http string) *HttpServer {
	return &HttpServer{http: http}
}

func writeJson(ctx iris.Context, out interface{}) {
	if !commons.IsNil(out) {
		switch out.(type) {
		case error:
			ctx.StatusCode(iris.StatusInternalServerError)
			perr := protocol.ConvertError(out.(error))
			_, _ = ctx.JSON(struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			}{Code: perr.Code(), Message: perr.Message()})
		default:
			ctx.StatusCode(iris.StatusOK)
			_, _ = ctx.JSON(out)
		}
	} else {
		ctx.StatusCode(iris.StatusNoContent)
	}
}
