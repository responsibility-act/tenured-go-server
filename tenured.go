package main

import (
	"github.com/ihaiker/tenured-go-server/commons"
	"github.com/ihaiker/tenured-go-server/services/store"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:     "tenured",
	Short:   "Tenured A completely open source IM cloud system.",
	Long:    `Complete documentation is available at http://tenured.renzhen.la`,
	Version: "1.0.0",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Usage()
	},
}

func init() {
	rootCmd.AddCommand(store.StoreCmd)
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "debug module")
}

func initConfig() {
	if debug, err := rootCmd.PersistentFlags().GetBool("debug"); err != nil {
		os.Exit(1)
	} else if debug {
		commons.InitLogrus(logrus.DebugLevel)
	} else {
		commons.InitLogrus(logrus.InfoLevel)
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
