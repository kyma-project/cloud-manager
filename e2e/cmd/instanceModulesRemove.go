package main

import (
	"fmt"

	"github.com/elliotchance/pie/v2"
	e2ekeb "github.com/kyma-project/cloud-manager/e2e/keb"
	"github.com/kyma-project/cloud-manager/pkg/external/operatorv1beta2"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"
)

type cmdInstanceModulesRemoveOptionsType struct {
	runtimeID  string
	alias      string
	moduleName string
}

var cmdInstanceModulesRemoveOptions cmdInstanceModulesRemoveOptionsType

var cmdInstanceModulesRemove = &cobra.Command{
	Use: "remove",
	RunE: func(cmd *cobra.Command, args []string) error {
		keb, err := e2ekeb.Create(rootCtx, config)
		if err != nil {
			return fmt.Errorf("failed to create keb: %w", err)
		}

		if cmdInstanceModulesRemoveOptions.runtimeID == "" {
			idArr, err := keb.List(rootCtx, e2ekeb.WithAlias(cmdInstanceModulesRemoveOptions.alias))
			if err != nil {
				return fmt.Errorf("failed to list runtimes: %w", err)
			}
			if len(idArr) == 0 {
				return fmt.Errorf("runtime with alias %q not found", cmdInstanceModulesRemoveOptions.alias)
			}
			if len(idArr) > 1 {
				return fmt.Errorf("multiple runtimes with alias %q found: %v", cmdInstanceModulesRemoveOptions.alias, pie.Map(idArr, func(x e2ekeb.InstanceDetails) string {
					return x.RuntimeID
				}))
			}
			cmdInstanceModulesRemoveOptions.runtimeID = idArr[0].RuntimeID
		}

		clnt, err := keb.CreateInstanceClient(rootCtx, cmdInstanceModulesRemoveOptions.runtimeID)
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
			if m.Name == cmdInstanceModulesRemoveOptions.moduleName {
				isFound = true
				break
			}
		}

		if !isFound {
			fmt.Println("Module is already removed")
			return nil
		}

		kyma.Spec.Modules = pie.FilterNot(kyma.Spec.Modules, func(m operatorv1beta2.Module) bool {
			return m.Name == cmdInstanceModulesRemoveOptions.moduleName
		})

		err = clnt.Update(rootCtx, kyma)
		if err != nil {
			return fmt.Errorf("failed to update SKR kyma: %w", err)
		}

		fmt.Println("Module is removed")

		return nil
	},
}

func init() {
	cmdInstanceModules.AddCommand(cmdInstanceModulesRemove)
	cmdInstanceModulesRemove.Flags().StringVarP(&cmdInstanceModulesRemoveOptions.runtimeID, "runtime-id", "r", "", "The runtime ID")
	cmdInstanceModulesRemove.Flags().StringVarP(&cmdInstanceModulesRemoveOptions.alias, "alias", "a", "", "The runtime alias")
	cmdInstanceModulesRemove.Flags().StringVarP(&cmdInstanceModulesRemoveOptions.moduleName, "module", "m", "", "The module name")
	_ = cmdInstanceModulesRemove.MarkFlagRequired("module")
	cmdInstanceModulesRemove.MarkFlagsMutuallyExclusive("runtime-id", "alias")
	cmdInstanceModulesRemove.MarkFlagsOneRequired("runtime-id", "alias")
}
