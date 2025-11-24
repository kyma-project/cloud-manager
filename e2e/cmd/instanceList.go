package main

import (
	"encoding/json"
	"fmt"

	e2ekeb "github.com/kyma-project/cloud-manager/e2e/keb"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
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

		if outputFormat == "json" {
			txt, err := json.MarshalIndent(arr, "", "  ")
			if err != nil {
				return fmt.Errorf("error marshalling json: %w", err)
			}
			fmt.Println(string(txt))
			return nil
		}

		if outputFormat == "yaml" {
			txt, err := yaml.Marshal(arr)
			if err != nil {
				return fmt.Errorf("error marshalling yaml: %w", err)
			}
			fmt.Println(string(txt))
			return nil
		}

		fmt.Println("")
		if len(arr) == 0 {
			fmt.Println("No instances found")
		} else {
			var tbl table.Table
			if outputFormat == "wide" {
				tbl = table.New("Alias", "RuntimeID", "Shoot", "Provider", "Region", "GA", "SA", "Ready", "Deleting")
			} else {
				tbl = table.New("Alias", "RuntimeID", "Shoot", "Provider", "Region", "Ready", "Deleting")
			}

			for _, id := range arr {
				if outputFormat == "wide" {
					tbl.AddRow(
						id.Alias,
						id.RuntimeID,
						id.ShootName,
						id.Provider,
						id.Region,
						id.GlobalAccount,
						id.SubAccount,
						id.ProvisioningCompleted,
						id.BeingDeleted,
					)
				} else {
					tbl.AddRow(
						id.Alias,
						id.RuntimeID,
						id.ShootName,
						id.Provider,
						id.Region,
						id.ProvisioningCompleted,
						id.BeingDeleted,
					)
				}
			}
			tbl.Print()
		}
		fmt.Println("")

		return nil
	},
}

func init() {
	cmdInstance.AddCommand(cmdInstanceList)
	cmdInstanceList.Flags().StringVarP(&outputFormat, "output", "o", "default", "Output format, one of: default, wide, json, yaml")
}
