package console

import (
	"github.com/ihaiker/tenured-go-server/services/console/controller"
)

type ConsoleServer struct {
	config     *ConsoleConfig
	httpServer *ctl.HttpServer
}

func (this *ConsoleServer) Start() error {
	logger.Info("start console http server")
	return this.httpServer.Start()
}

func (this *ConsoleServer) Shutdown(interrupt bool) {
	logger.Info("stop console http server")
	this.httpServer.Shutdown(interrupt)
}

func newConsoleServer(config *ConsoleConfig) (*ConsoleServer, error) {
	server := &ConsoleServer{
		config:     config,
		httpServer: ctl.NewHttpServer(config.HTTP),
	}
	return server, nil
}
