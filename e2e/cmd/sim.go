package main

import "github.com/spf13/cobra"

var cmdSim = &cobra.Command{
	Use: "sim",
}

func init() {
	cmdRoot.AddCommand(cmdSim)
}
