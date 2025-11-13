package main

import (
	"fmt"

	e2ekeb "github.com/kyma-project/cloud-manager/e2e/keb"
	"github.com/spf13/cobra"
)

var cmdInstanceList = &cobra.Command{
	Use:   "list",
	Short: "List all instances",
	RunE: func(cmd *cobra.Command, args []string) error {
		keb, err := e2ekeb.Create(rootCtx, config)
		if err != nil {
			return fmt.Errorf("failed to create keb: %w", err)
		}

		arr, err := keb.List(rootCtx)
		if err != nil {
			return fmt.Errorf("error listing intances: %w", err)
		}

		fmt.Println("")
		if len(arr) == 0 {
			fmt.Println("No instances found")
		} else {
			tpl := "%-17s %-37s %-9s %-10s %-10s %-37s %-37s\n"
			fmt.Printf(
				tpl,
				"Alias",
				"RuntimeID",
				"Shoot",
				"Provider",
				"Region",
				"GA",
				"SA",
			)
			for _, id := range arr {
				fmt.Printf(
					tpl,
					id.Alias,
					id.RuntimeID,
					id.ShootName,
					id.Provider,
					id.Region,
					id.GlobalAccount,
					id.SubAccount,
				)
			}
		}
		fmt.Println("")

		return nil
	},
}

func init() {
	cmdInstance.AddCommand(cmdInstanceList)
}
