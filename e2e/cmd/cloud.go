package main

import "github.com/spf13/cobra"

var cmdCloud = &cobra.Command{
	Use: "cloud",
}

func init() {
	cmdRoot.AddCommand(cmdCloud)
}
