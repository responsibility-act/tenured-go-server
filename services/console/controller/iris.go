package ctl

import (
	"context"
	"github.com/ihaiker/tenured-go-server/protocol"
	"github.com/ihaiker/tenured-go-server/registry/load_balance"

	"github.com/ihaiker/tenured-go-server/api"
	"github.com/ihaiker/tenured-go-server/api/client"
	"github.com/ihaiker/tenured-go-server/commons"
	"github.com/ihaiker/tenured-go-server/commons/logs"
	"github.com/kataras/iris/v12"
	ctx "github.com/kataras/iris/v12/context"
	"time"
)

var app = iris.Default()
var logger = logs.GetLogger("iris")

var accountService api.AccountService
var clusterIdService api.ClusterIdService
var userService api.UserService

type HttpServer struct {
	http           string
	serviceManager *commons.ServiceManager
}

func allService() []interface{} {
	return []interface{}{accountService, clusterIdService, userService}
}

func (this *HttpServer) startService() (err error) {
	for _, s := range allService() {
		if err = commons.StartIfService(s); err != nil {
			return
		}
	}
	return nil
}

func (this *HttpServer) Start() (err error) {
	logger.Debug("http server start: ", this.http)
	if err = this.startService(); err != nil {
		return
	}
	app.Logger().SetLevel(logger.Level.String())
	app.Logger().SetOutput(logger.Out)
	app.Logger().SetTimeFormat("2006-01-02 15:04:05")
	app.Logger().SetPrefix("(iris) ")
	app.OnErrorCode(iris.StatusNotFound, func(ctx iris.Context) {
		writeJson(ctx, protocol.NewError("404", "NotFound"))
	})
	app.OnErrorCode(iris.StatusInternalServerError, func(ctx iris.Context) {
		writeJson(ctx, protocol.NewError("502", "InternalServerError"))
	})

	app.Get("/health", func(ctx ctx.Context) {
		ctx.JSON(map[string]interface{}{"status": "UP"})
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

func (this *HttpServer) shutdownService(interrupt bool) {
	for _, s := range allService() {
		commons.ShutdownIfService(s, interrupt)
	}
}

func (this *HttpServer) Shutdown(interrupt bool) {
	this.shutdownService(interrupt)

	if err := app.Shutdown(context.Background()); err != nil {
		logger.Error("shutdown console http server error:", err)
	}
}

func NewHttpServer(http string, storeClientLoadBalance load_balance.LoadBalance) *HttpServer {
	accountService = client.NewAccountServiceClient(storeClientLoadBalance)
	clusterIdService = client.NewClusterIdServiceClient(storeClientLoadBalance)
	userService = client.NewUserServiceClient(storeClientLoadBalance)

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
