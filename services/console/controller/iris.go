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
	log, _ = logs.GetLogger("console")

	requestLogger := logger.New(logger.Config{
		// Status displays status code
		Status: true,
		// IP displays request's remote address
		IP: true,
		// Method displays the http method
		Method: true,
		// Path displays the request path
		Path: true,
		// Query appends the url query to the Path.
		Query: true,
		// if !empty then its contents derives from `ctx.Values().Get("logger_message")
		// will be added to the logs.
		//MessageContextKeys: []string{"logger_message"},
		// if !empty then its contents derives from `ctx.GetHeader("User-Agent")
		//MessageHeaderKeys: []string{"User-Agent"},
	})
	app.Use(requestLogger)

	app.Get("/health", func(ctx ctx.Context) {
		ctx.JSON(map[string]interface{}{"status": "UP"})
	})
}
