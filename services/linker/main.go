package linker

import (
	"github.com/ihaiker/tenured-go-server/commons/logs"
	"github.com/ihaiker/tenured-go-server/commons/runtime/signal"
	"github.com/ihaiker/tenured-go-server/services"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var logger *logrus.Logger
var linkerService *LinkerServer
var linkerCfg *linkerConfig

var LinkerCmd = &cobra.Command{
	Use:   "linker",
	Short: "Tenured Linker Server",
	Long:  `Complete documentation is available at http://tenured.renzhen.la/linker`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		config, err := cmd.PersistentFlags().GetString("config")
		if err != nil {
			return err
		}
		linkerCfg = NewLinkerConfig()
		if err := services.LoadServerConfig("linker", config, linkerCfg); err != nil {
			return err
		}

		if err = logs.InitLogger(
			linkerCfg.Logs.Loggers,
			linkerCfg.Logs.Level, linkerCfg.Logs.Output, linkerCfg.Logs.Path,
			linkerCfg.Logs.Archive,
		); err != nil {
			return err
		}
		logger = logs.GetLogger("linker")
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		linkerService = NewLinkerServer(linkerCfg)
		err := linkerService.Start()
		if err == nil {
			signal.Signal(func() {})
		} else {
			logger.Error(err.Error())
		}
		return err
	},
	PostRun: func(cmd *cobra.Command, args []string) {
		if linkerService != nil {
			linkerService.Shutdown(false)
		}
	},
}

func init() {
	LinkerCmd.PersistentFlags().StringP("config", "f", "",
		`the config file. 
default: ${workDir}/conf/linker.{yaml|json} or /etc/tenured/linker.{yaml|json}`)

}
