package store

import (
	"github.com/ihaiker/tenured-go-server/commons/runtime"
	"github.com/ihaiker/tenured-go-server/commons/runtime/signal"
	"github.com/kataras/iris/core/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"
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
		if config != "" {
			storeCfg, err = initConfig(config)
			return err
		} else {
			searchConfigs := []string{
				runtime.GetWorkDir() + "/conf/store.yml",
				runtime.GetWorkDir() + "/conf/store.json",
				"/etc/tenured/conf/store.yml",
				"/etc/tenured/conf/store.json",
			}
			for _, searchConfig := range searchConfigs {
				if storeCfg, err = initConfig(searchConfig); err == nil {
					logrus.Info("use config file: ", searchConfig)
					return nil
				} else {
					logrus.Debugf("search config file %s not found!", searchConfig)
				}
			}
			return errors.New("any config found ! \n\t" + strings.Join(searchConfigs, "\n\t"))
		}
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
