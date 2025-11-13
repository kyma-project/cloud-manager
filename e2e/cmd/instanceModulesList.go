package main

import (
	"fmt"

	e2ekeb "github.com/kyma-project/cloud-manager/e2e/keb"
	"github.com/kyma-project/cloud-manager/pkg/external/operatorv1beta2"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"
)

var cmdInstanceModulesList = &cobra.Command{
	Use: "list",
	RunE: func(cmd *cobra.Command, args []string) error {
		keb, err := e2ekeb.Create(rootCtx, config)
		if err != nil {
			return fmt.Errorf("failed to create keb: %w", err)
		}

		kyma := &operatorv1beta2.Kyma{}
		err = keb.KcpClient().Get(rootCtx, types.NamespacedName{
			Namespace: config.KcpNamespace,
			Name:      runtimeID,
		}, kyma)
		if err != nil {
			return fmt.Errorf("failed to get KCP kyma: %w", err)
		}

		fmt.Println("")
		tpl := "%-10s %-10s %s\n"
		fmt.Printf(tpl, "Module", "State", "Message")
		for _, m := range kyma.Status.Modules {
			fmt.Printf(
				tpl,
				m.Name,
				m.State,
				m.Message,
			)
		}

		return nil
	},
}

func init() {
	cmdInstanceModules.AddCommand(cmdInstanceModulesList)
	cmdInstanceModulesList.Flags().StringVarP(&runtimeID, "runtime-id", "r", "", "The runtime ID")
	_ = cmdInstanceModulesList.MarkFlagRequired("runtime-id")
}
