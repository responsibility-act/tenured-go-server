package tenant

import (
	"github.com/ihaiker/tenured-go-server/commons/logs"
	"github.com/ihaiker/tenured-go-server/commons/runtime/signal"
	"github.com/ihaiker/tenured-go-server/services"
	"github.com/spf13/cobra"
)

var logger = logs.GetLogger("console")
var tenantServer *TenantServer
var tenantConfig *TenantConfig

var TenantCommand = &cobra.Command{
	Use:   "tenant",
	Short: "tenant restful api",
	Long:  `Complete documentation is available at http://tenured.renzhen.la/tenant.html`,
	//Args:    cobra.MinimumNArgs(1),
	Example: `	tenured tenant -f <path>`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		config, err := cmd.PersistentFlags().GetString("config")
		if err != nil {
			return err
		}
		tenantConfig = NewTenantConfig()
		if err := services.LoadServerConfig("tenant", config, tenantConfig); err != nil {
			return err
		}

		if err = logs.InitLogger(
			tenantConfig.Logs.Loggers,
			tenantConfig.Logs.Level, tenantConfig.Logs.Output, tenantConfig.Logs.Path,
			tenantConfig.Logs.Archive,
		); err != nil {
			return err
		}

		return err
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		tenantServer, err = newTenantServer(tenantConfig)
		if err != nil {
			return
		}
		err = tenantServer.Start()
		if err == nil {
			signal.Signal(func() {})
		} else {
			logger.Error(err.Error())
		}
		return err
	},
	PostRun: func(cmd *cobra.Command, args []string) {
		if tenantServer != nil {
			tenantServer.Shutdown(false)
		}
	},
}

func init() {
	TenantCommand.PersistentFlags().StringP("config", "f", "",
		`the config file. 
default: ${workDir}/conf/tenant.{yaml|json} or /etc/tenured/tenant.{yaml|json}`)
}
