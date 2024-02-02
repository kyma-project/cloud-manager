package main

import "github.com/spf13/cobra"

func init() {
	cmdKymaModule.AddCommand(cmdKymaModuleState)
}

var cmdKymaModuleState = &cobra.Command{
	Use:     "state",
	Aliases: []string{"s"},
}
