package main

import "github.com/spf13/cobra"

var cmdManifest = &cobra.Command{
	Use: "manifest",
}

func init() {
	cmdRoot.AddCommand(cmdManifest)
}
