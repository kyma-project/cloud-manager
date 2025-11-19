package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

var cmdConfigDump = &cobra.Command{
	Use:   "dump",
	Short: "Prints the loaded config",
	RunE: func(cmd *cobra.Command, args []string) error {
		x, err := yaml.Marshal(config)
		if err != nil {
			return err
		}
		fmt.Println(string(x))
		return nil
	},
}

func init() {
	cmdConfig.AddCommand(cmdConfigDump)
}
