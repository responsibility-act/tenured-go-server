package store

import (
	"github.com/ihaiker/tenured-go-server/commons/logs"
	"github.com/ihaiker/tenured-go-server/commons/runtime/signal"
	"github.com/ihaiker/tenured-go-server/services"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var logger *logrus.Logger
var storeService *storeServer
var storeCfg *storeConfig

var StoreCmd = &cobra.Command{
	Use:     "store",
	Short:   "Tenured Store Server",
	Long:    `Complete documentation is available at http://tenured.renzhen.la/store`,
	Version: "1.0.0",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		config, err := cmd.PersistentFlags().GetString("config")
		if err != nil {
			return err
		}
		storeCfg = NewStoreConfig()
		if err := services.LoadServerConfig("store", config, storeCfg); err != nil {
			return err
		}

		if err = os.Chdir(storeCfg.WorkDir); err != nil {
			return err
		}

		if debug, err := cmd.Root().PersistentFlags().GetBool("debug"); err == nil && debug {
			storeCfg.Logs.Level = "debug"
		}

		if logger, err = logs.InitLogger(
			"store",
			storeCfg.Logs.Output, storeCfg.Logs.Level,
			storeCfg.Logs.Path, storeCfg.Logs.Archive,
		); err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		storeService = newStoreServer(storeCfg)
		err = storeService.Start()
		if err == nil {
			signal.Signal(func() {
				//do nothing...
			})
		}
		return
	},
	PostRun: func(cmd *cobra.Command, args []string) {
		if storeService != nil {
			storeService.Shutdown(false)
		}
	},
}

func init() {
	StoreCmd.PersistentFlags().StringP("config", "f", "",
		`the config file. 
default: ${workDir}/conf/store.{yaml|json} or /etc/tenured/store.{yaml|json}`)

}
