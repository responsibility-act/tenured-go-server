package console

import (
	"context"
	"github.com/ihaiker/tenured-go-server/services/console/controller"
	"github.com/kataras/iris"
	"github.com/sirupsen/logrus"
)

type ConsoleServer struct {
	config *ConsoleConfig
	app    *iris.Application
}

func (this *ConsoleServer) Start() error {
	logrus.Info("start console ...")
	if app, err := controller.StartIris(this.config.HTTP); err != nil {
		return err
	} else {
		this.app = app
		return nil
	}
}

func (this *ConsoleServer) Shutdown(interrupt bool) {
	logrus.Info("stop console ...")
	if err := this.app.Shutdown(context.Background()); err != nil {
		logrus.Error("close iris web error: %s", err)
	}
}

func newConsoleServer(config *ConsoleConfig) (*ConsoleServer, error) {
	server := &ConsoleServer{
		config: config,
	}
	return server, nil
}
