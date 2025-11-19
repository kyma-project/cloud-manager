package main

import (
	"fmt"

	e2ekeb "github.com/kyma-project/cloud-manager/e2e/keb"
	"github.com/kyma-project/cloud-manager/pkg/external/operatorv1beta2"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"
)

var cmdInstanceModulesAdd = &cobra.Command{
	Use: "add",
	RunE: func(cmd *cobra.Command, args []string) error {
		keb, err := e2ekeb.Create(rootCtx, config)
		if err != nil {
			return fmt.Errorf("failed to create keb: %w", err)
		}

		clnt, err := keb.CreateInstanceClient(rootCtx, runtimeID)
		if err != nil {
			return err
		}

		kyma := &operatorv1beta2.Kyma{}
		err = clnt.Get(rootCtx, types.NamespacedName{
			Namespace: "kyma-system",
			Name:      "default",
		}, kyma)
		if err != nil {
			return fmt.Errorf("failed to get SKR kyma: %w", err)
		}

		isFound := false
		for _, m := range kyma.Spec.Modules {
			if m.Name == moduleName {
				isFound = true
				break
			}
		}

		if isFound {
			fmt.Println("Module is already added")
			return nil
		}

		kyma.Spec.Modules = append(kyma.Spec.Modules, operatorv1beta2.Module{
			Name: moduleName,
		})

		err = clnt.Update(rootCtx, kyma)
		if err != nil {
			return fmt.Errorf("failed to update SKR kyma: %w", err)
		}

		fmt.Println("Module is added")

		return nil
	},
}

func init() {
	cmdInstanceModules.AddCommand(cmdInstanceModulesAdd)
	cmdInstanceModulesAdd.Flags().StringVarP(&runtimeID, "runtime-id", "r", "", "The runtime ID")
	cmdInstanceModulesAdd.Flags().StringVarP(&moduleName, "module", "m", "", "The module name")
	_ = cmdInstanceModulesAdd.MarkFlagRequired("runtime-id")
	_ = cmdInstanceModulesAdd.MarkFlagRequired("module")
}
