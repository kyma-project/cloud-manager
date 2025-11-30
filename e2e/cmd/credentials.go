package main

import "github.com/spf13/cobra"

var cmdCredentials = &cobra.Command{
	Use:     "credentials",
	Aliases: []string{"c", "cred", "creds"},
}

func init() {
	cmdRoot.AddCommand(cmdCredentials)
}
