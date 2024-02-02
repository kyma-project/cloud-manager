package main

import "github.com/spf13/cobra"

var (
	kymaName string
)

func init() {
	cmdKyma.PersistentFlags().StringVarP(&kymaName, "kymaName", "k", "", "Kyma CR name")
	cmdRoot.AddCommand(cmdKyma)
}

var cmdKyma = &cobra.Command{
	Use:     "kyma",
	Aliases: []string{"k"},
}
