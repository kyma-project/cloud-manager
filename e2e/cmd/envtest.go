package main

import "github.com/spf13/cobra"

var cmdEnvtest = &cobra.Command{
	Use: "envtest",
}

func init() {
	cmdRoot.AddCommand(cmdEnvtest)
}
