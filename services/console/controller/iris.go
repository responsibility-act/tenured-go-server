package ctl

import (
	"context"
	"github.com/ihaiker/tenured-go-server/commons/logs"
	"github.com/kataras/iris"
	ctx "github.com/kataras/iris/context"
	"github.com/kataras/iris/middleware/logger"
	"github.com/sirupsen/logrus"
)

var app *iris.Application = iris.Default()
var log *logrus.Logger

type HttpServer struct {
	http string
}

func (this *HttpServer) Start() error {
	log.Info("http server start: ", this.http)
	return app.Run(
		iris.Addr(this.http),
		iris.WithoutBanner,
		iris.WithoutServerError(iris.ErrServerClosed),
	)
}

func (this *HttpServer) Shutdown(interrupt bool) {
	if err := app.Shutdown(context.Background()); err != nil {
		log.Error("shutdown console http server error:", err)
	}
}

func NewHttpServer(http string) *HttpServer {
	return &HttpServer{http: http}
}

func init() {
	log = logs.GetLogger("console")

	loggerConfig := logger.DefaultConfig()
	loggerConfig.Query = true
	requestLogger := logger.New(loggerConfig)
	app.Use(requestLogger)

	app.Logger().SetOutput(log.Out)

	app.Get("/health", func(ctx ctx.Context) {
		ctx.JSON(map[string]interface{}{"status": "UP"})
	})
}
