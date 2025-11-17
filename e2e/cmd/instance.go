package main

import (
	"github.com/spf13/cobra"
)

var cmdInstance = &cobra.Command{
	Use: "instance",
}

func init() {
	cmdRoot.AddCommand(cmdInstance)
}
