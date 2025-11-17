package main

import (
	"fmt"

	e2ekeb "github.com/kyma-project/cloud-manager/e2e/keb"
	"github.com/rodaine/table"
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
			tbl := table.New("Alias", "RuntimeID", "Shoot", "Provider", "Region", "GA", "SA", "Ready")

			for _, id := range arr {
				tbl.AddRow(
					id.Alias,
					id.RuntimeID,
					id.ShootName,
					id.Provider,
					id.Region,
					id.GlobalAccount,
					id.SubAccount,
					id.ProvisioningCompleted,
				)
			}
			tbl.Print()
		}
		fmt.Println("")

		return nil
	},
}

func init() {
	cmdInstance.AddCommand(cmdInstanceList)
}
