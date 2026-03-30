package main

import (
	"fmt"

	"github.com/elliotchance/pie/v2"
	e2ekeb "github.com/kyma-project/cloud-manager/e2e/keb"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type cmdInstanceShowOptionsType struct {
	runtimeID string
	alias     string
}

var cmdInstanceShowOptions cmdInstanceShowOptionsType

var cmdInstanceShow = &cobra.Command{
	Use:   "show",
	Short: "Show instance details",
	RunE: func(cmd *cobra.Command, args []string) error {
		keb, err := e2ekeb.Create(rootCtx, config)
		if err != nil {
			return fmt.Errorf("failed to create keb: %w", err)
		}

		var id e2ekeb.InstanceDetails
		if cmdInstanceShowOptions.runtimeID != "" {
			x, err := keb.GetInstance(rootCtx, cmdInstanceShowOptions.runtimeID)
			if err != nil {
				return fmt.Errorf("failed to get instance: %w", err)
			}
			if x == nil {
				return fmt.Errorf("instance %q not found", cmdInstanceShowOptions.runtimeID)
			}
			id = *x
		} else {
			arr, err := keb.List(rootCtx, e2ekeb.WithAlias(cmdInstanceShowOptions.alias))
			if err != nil {
				return fmt.Errorf("failed to list instances by alias: %w", err)
			}
			if len(arr) == 0 {
				return fmt.Errorf("instance %q not found", cmdInstanceShowOptions.alias)
			}
			if len(arr) > 1 {
				return fmt.Errorf("more than one instance found: %v", pie.Map(arr, func(xx e2ekeb.InstanceDetails) string {
					return xx.RuntimeID
				}))
			}
			id = arr[0]
		}

		b, err := yaml.Marshal(id)
		if err != nil {
			return fmt.Errorf("failed to marshal instance: %w", err)
		}
		fmt.Println(string(b))

		return nil
	},
}

func init() {
	cmdInstance.AddCommand(cmdInstanceShow)
	cmdInstanceShow.Flags().StringVarP(&cmdInstanceShowOptions.runtimeID, "runtime-id", "r", "", "Runtime ID")
	cmdInstanceShow.Flags().StringVarP(&cmdInstanceShowOptions.alias, "alias", "a", "", "Alias name")
	cmdInstanceShow.MarkFlagsMutuallyExclusive("runtime-id", "alias")
	cmdInstanceShow.MarkFlagsOneRequired("runtime-id", "alias")
}
