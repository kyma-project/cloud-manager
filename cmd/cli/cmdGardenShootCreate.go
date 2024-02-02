package main

import (
	"github.com/spf13/cobra"
)

func init() {
	cmdGardenShoot.AddCommand(cmdGardenShootCreate)
}

var cmdGardenShootCreate = &cobra.Command{
	Use:     "create",
	Aliases: []string{"c"},
}
