package main

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/kyma-project/cloud-manager/cmd/cli/helper"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"github.com/spf13/cobra"
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

		kyma := util.NewKymaUnstructured()

		if kymaName == "" {
			fmt.Println("Kyma name not specified it will be generated.")
			kymaName = uuid.NewString()
		}
		kyma.SetName(kymaName)
		kyma.SetNamespace("kcp-system")

		if shootName != "" {
			fmt.Println("Adding kyma-project.io/shoot-name label")
			kyma.SetLabels(map[string]string{"kyma-project.io/shoot-name": shootName})
		}

		err := c.Create(context.Background(), kyma)

		if err != nil {
			return fmt.Errorf("error creating Kyma: %w", err)
		}

		fmt.Printf("Created Kyma %v\n", kymaName)
		return nil
	},
}
