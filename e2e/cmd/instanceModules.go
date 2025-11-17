package main

import "github.com/spf13/cobra"

var cmdInstanceModules = &cobra.Command{
	Use:     "modules",
	Aliases: []string{"mo", "mod", "mods", "module"},
}

func init() {
	cmdInstance.AddCommand(cmdInstanceModules)
}
