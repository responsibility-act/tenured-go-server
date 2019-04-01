package tools

import (
	"github.com/ihaiker/tenured-go-server/commons/mixins"
	"github.com/spf13/cobra"
)

var InstallCommand = &cobra.Command{
	Use:   "install",
	Short: "Automatically initialize environment dependencies",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	registry := mixins.Get(mixins.KeyRegistry, mixins.Registry)
	InstallCommand.PersistentFlags().StringP("registry", "g", registry, "the registry server url.")
}
