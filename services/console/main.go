package console

import "C"
import (
	"github.com/ihaiker/tenured-go-server/commons/logs"
	"github.com/ihaiker/tenured-go-server/services"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var logger *logrus.Logger
var consoleServer *ConsoleServer
var consoleConfig *ConsoleConfig

var ConsoleCommand = &cobra.Command{
	Use:     "console",
	Short:   "Tenured Console",
	Long:    `Complete documentation is available at http://tenured.renzhen.la/console`,
	Version: "1.0.0",
	//Args:    cobra.MinimumNArgs(1),
	Example: `	tenured console -f <path>`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		config, err := cmd.PersistentFlags().GetString("config")
		if err != nil {
			return err
		}
		consoleConfig = NewConsoleConfig()
		if err := services.LoadServerConfig("console", config, consoleConfig); err != nil {
			return err
		}
		if err = os.Chdir(consoleConfig.WorkDir); err != nil {
			return err
		}

		if debug, err := cmd.Root().PersistentFlags().GetBool("debug"); err == nil && debug {
			consoleConfig.Logs.Level = "debug"
		}

		logger, err = logs.InitLogger(
			"console",
			consoleConfig.Logs.Output,
			consoleConfig.Logs.Level,
			consoleConfig.Logs.Path,
			consoleConfig.Logs.Archive,
		)
		return err
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		consoleServer, err = newConsoleServer(consoleConfig)
		if err != nil {
			return
		}
		return consoleServer.Start()
	},
	PostRun: func(cmd *cobra.Command, args []string) {
		if consoleServer != nil {
			consoleServer.Shutdown(false)
		}
	},
}

func init() {
	ConsoleCommand.PersistentFlags().StringP("config", "f", "",
		`the config file. 
default: ${workDir}/conf/console.{yaml|json} or /etc/tenured/console.{yaml|json}`)
}
