package main

import "github.com/spf13/cobra"

func init() {
	cmdRoot.AddCommand(cmdGarden)
}

var cmdGarden = &cobra.Command{
	Use:     "garden",
	Aliases: []string{"g"},
}
