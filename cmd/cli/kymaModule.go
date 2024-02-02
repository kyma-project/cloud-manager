package main

import "github.com/spf13/cobra"

var (
	kymaModuleName string = "cloud-manager"
)

func init() {
	cmdKymaModule.PersistentFlags().StringVarP(&kymaModuleName, "module", "m", "cloud-manager", "Kyma module name, defaults to cloud-manager")
	cmdKyma.AddCommand(cmdKymaModule)
}

var cmdKymaModule = &cobra.Command{
	Use:     "module",
	Aliases: []string{"m"},
}
