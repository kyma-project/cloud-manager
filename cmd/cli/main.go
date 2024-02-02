package main

import (
	"github.com/spf13/cobra"
)

var (
	namespace string
)

func init() {
	cmdRoot.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "Namespace")
}

var cmdRoot = &cobra.Command{
	Use: "cli",
}

func main() {
	err := cmdRoot.Execute()
	if err != nil {
		panic(err)
	}
}
