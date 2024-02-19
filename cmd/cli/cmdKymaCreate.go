package main

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/cmd/cli/helper"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"github.com/spf13/cobra"
	"strings"
)

func init() {
	cmdKymaCreate.Flags().StringVarP(&shootName, "shoot", "s", "", "Shoot name")

	cmdKyma.AddCommand(cmdKymaCreate)
}

var cmdKymaCreate = &cobra.Command{
	Use:     "create",
	Aliases: []string{"c"},
	RunE: func(cmd *cobra.Command, args []string) error {
		c := helper.NewKcpClient()

		list := util.NewKymaListUnstructured()
		err := c.List(context.Background(), list)
		if err != nil {
			return fmt.Errorf("error listing Kymas: %w", err)
		}

		fmt.Printf("KymaName \t\t\t\t\t\t Module %s state\n", kymaModuleName)
		for _, item := range list.Items {
			state := util.GetKymaModuleStateFromStatus(&item, kymaModuleName)
			name := item.GetName()
			if kymaName == "" || strings.Contains(name, kymaName) {
				fmt.Printf("%s/%s \t %s\n", item.GetNamespace(), name, state)
			}
		}

		return nil
	},
}
