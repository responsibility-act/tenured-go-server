package store

import (
	"github.com/ihaiker/tenured-go-server/commons/logs"
	"github.com/ihaiker/tenured-go-server/commons/runtime/signal"
	"github.com/ihaiker/tenured-go-server/services"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

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
		return services.LoadServerConfig("store", config, storeCfg)
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		if err = os.Chdir(storeCfg.WorkDir); err != nil {
			return
		}

		if debug, err := cmd.Root().PersistentFlags().GetBool("debug"); err == nil && debug {
			storeCfg.Logs.Level = "debug"
		}
		if err = logs.InitLogrus(storeCfg.Logs.Output, storeCfg.Logs.Level,
			storeCfg.Logs.Path, storeCfg.Logs.Archive); err != nil {
			return err
		}
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

func logurs(agent string) *logrus.Entry {
	return logrus.WithField("agent", agent)
}
