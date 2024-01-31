package main

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "cli",
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
}
