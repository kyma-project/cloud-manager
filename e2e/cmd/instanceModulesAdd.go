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

		// TODO continue here

		return fmt.Errorf("not implemented")
	},
}

func init() {
	cmdInstanceModules.AddCommand(cmdInstanceModulesAdd)
	cmdInstanceModulesAdd.Flags().StringVarP(&runtimeID, "runtime-id", "r", "", "The runtime ID")
	cmdInstanceModulesAdd.Flags().StringVarP(&moduleName, "module", "m", "", "The module name")
	_ = cmdInstanceModulesList.MarkFlagRequired("runtime-id")
	_ = cmdInstanceModulesList.MarkFlagRequired("module")
}
