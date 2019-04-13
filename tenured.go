package main

import (
	"fmt"
	"github.com/ihaiker/tenured-go-server/commons/logs"
	"github.com/ihaiker/tenured-go-server/services/console"
	"github.com/ihaiker/tenured-go-server/services/store"
	"github.com/ihaiker/tenured-go-server/tools"
	"github.com/spf13/cobra"
	"os"
	"runtime"
)

var (
	VERSION    string
	BUILD_TIME string
	GO_VERSION string
)

var rootCmd = &cobra.Command{
	Use: "tenured",
	Short: "Tenured A completely open source IM cloud system.\n" +
		"Build: " + BUILD_TIME + "\n" +
		GO_VERSION,
	Long:    `Complete documentation is available at http://tenured.renzhen.la`,
	Version: VERSION,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(cmd.Short)
		_ = cmd.Usage()
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if debug, err := cmd.Root().PersistentFlags().GetBool("debug"); err == nil && debug {
			logs.DebugLogger()
		}
	},
}

func init() {
	rootCmd.AddCommand(store.StoreCmd)
	rootCmd.AddCommand(console.ConsoleCommand)
	rootCmd.AddCommand(tools.ConfigCmd)
	rootCmd.AddCommand(tools.InstallCommand)
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "debug module")
}

func initConfig() {

}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
