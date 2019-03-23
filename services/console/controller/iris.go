package controller

import (
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/logger"
	"time"
)

func StartIris(http string) (*iris.Application, error) {
	app := iris.Default()
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

	c := make(chan error, 0)
	defer close(c)
	go func() {
		if err := app.Run(
			iris.Addr(http),
			//iris.WithoutBanner,
			iris.WithoutServerError(iris.ErrServerClosed),
		); err != nil {
			c <- err
		}
	}()

	select {
	case err := <-c:
		return app, err
	case <-time.After(time.Second):
		return app, nil
	}
}
