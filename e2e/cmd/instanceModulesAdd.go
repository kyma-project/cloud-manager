package main

import (
	"context"
	"fmt"
	"time"

	"github.com/elliotchance/pie/v2"
	"github.com/go-logr/logr"
	e2ekeb "github.com/kyma-project/cloud-manager/e2e/keb"
	"github.com/kyma-project/cloud-manager/pkg/external/operatorshared"
	"github.com/kyma-project/cloud-manager/pkg/external/operatorv1beta2"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
)

var cmdInstanceModulesAdd = &cobra.Command{
	Use: "add",
	RunE: func(cmd *cobra.Command, args []string) error {
		keb, err := e2ekeb.Create(rootCtx, config)
		if err != nil {
			return fmt.Errorf("failed to create keb: %w", err)
		}

		if runtimeID == "" {
			idArr, err := keb.List(rootCtx, e2ekeb.WithAlias(alias))
			if err != nil {
				return fmt.Errorf("failed to list runtimes: %w", err)
			}
			if len(idArr) == 0 {
				return fmt.Errorf("runtime with alias %q not found", alias)
			}
			if len(idArr) > 1 {
				return fmt.Errorf("multiple runtimes with alias %q found: %v", alias, pie.Map(idArr, func(x e2ekeb.InstanceDetails) string {
					return x.RuntimeID
				}))
			}
			runtimeID = idArr[0].RuntimeID
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

		if waitDone && moduleName == "cloud-manager" {
			fmt.Println("Waiting for module to be ready")

			logger := logr.Discard()
			if verbose {
				logger = rootLogger.WithName("waitModuleReady")
			}
			err := wait.PollUntilContextTimeout(rootCtx, 5*time.Second, timeout, false, func(ctx context.Context) (done bool, err error) {
				err = clnt.Get(rootCtx, types.NamespacedName{
					Namespace: "kyma-system",
					Name:      "default",
				}, kyma)
				if err != nil {
					return false, fmt.Errorf("failed to get SKR kyma while waiting module ready: %w", err)
				}
				m, ok := kyma.GetModuleStatusMap()[moduleName]
				if !ok {
					logger.Info("not found in status")
					return false, nil
				}
				if m.State != operatorshared.StateReady {
					logger.WithValues("state", m.State).Info("module is not ready")
					return false, nil
				}
				return true, nil
			})
			if err != nil {
				return fmt.Errorf("failed to wait for module ready: %w", err)
			}
		}

		return nil
	},
}

func init() {
	cmdInstanceModules.AddCommand(cmdInstanceModulesAdd)
	cmdInstanceModulesAdd.Flags().StringVarP(&runtimeID, "runtime-id", "r", "", "The runtime ID")
	cmdInstanceModulesAdd.Flags().StringVarP(&alias, "alias", "a", "", "The runtime alias")
	cmdInstanceModulesAdd.Flags().StringVarP(&moduleName, "module", "m", "", "The module name")
	cmdInstanceModulesAdd.Flags().BoolVarP(&waitDone, "wait", "w", false, "Wait until module is added")
	cmdInstanceModulesAdd.Flags().DurationVarP(&timeout, "timeout", "t", 5*time.Minute, "Timeout")
	_ = cmdInstanceModulesAdd.MarkFlagRequired("module")
	cmdInstanceModulesAdd.MarkFlagsMutuallyExclusive("runtime-id", "alias")
	cmdInstanceModulesAdd.MarkFlagsOneRequired("runtime-id", "alias")
}
