package console

import (
	"github.com/ihaiker/tenured-go-server/commons/registry"
)

func Server(plugins registry.RegistryPlugins, configFile string) (*ConsoleServer, error) {
	config := &ConsoleConfig{
		HTTP: ":2001",
	}
	return newConsoleServer(config)
}
