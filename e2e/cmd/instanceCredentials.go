package main

import "github.com/spf13/cobra"

var cmdInstanceCredentials = &cobra.Command{
	Use:     "credentials",
	Aliases: []string{"cre", "cred", "creds"},
}

func init() {
	cmdInstance.AddCommand(cmdInstanceCredentials)
}
