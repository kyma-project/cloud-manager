package main

import "github.com/spf13/cobra"

var (
	shootName string
)

func init() {
	cmdGardenShoot.PersistentFlags().StringVarP(&shootName, "shoot", "s", "", "Shoot name")
	cmdGarden.AddCommand(cmdGardenShoot)
}

var cmdGardenShoot = &cobra.Command{
	Use:     "shoot",
	Aliases: []string{"s"},
}
