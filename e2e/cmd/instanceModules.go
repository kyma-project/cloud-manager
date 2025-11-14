package main

import "github.com/spf13/cobra"

var moduleName string

var cmdInstanceModules = &cobra.Command{
	Use:     "modules",
	Aliases: []string{"mo", "mod", "mods", "module"},
}

func init() {
	cmdInstance.AddCommand(cmdInstanceModules)
}
