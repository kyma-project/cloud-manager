package main

import (
	"github.com/spf13/cobra"
)

var cmdConfig = &cobra.Command{
	Use: "config",
}

func init() {
	cmdRoot.AddCommand(cmdConfig)
}
