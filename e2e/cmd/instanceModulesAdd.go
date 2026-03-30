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

type cmdInstanceModulesAddOptionsType struct {
	runtimeID  string
	alias      string
	moduleName string
	waitDone   bool
	timeout    time.Duration
}

var cmdInstanceModulesAddOptions cmdInstanceModulesAddOptionsType

var cmdInstanceModulesAdd = &cobra.Command{
	Use: "add",
	RunE: func(cmd *cobra.Command, args []string) error {
		keb, err := e2ekeb.Create(rootCtx, config)
		if err != nil {
			return fmt.Errorf("failed to create keb: %w", err)
		}

		if cmdInstanceModulesAddOptions.runtimeID == "" {
			idArr, err := keb.List(rootCtx, e2ekeb.WithAlias(cmdInstanceModulesAddOptions.alias))
			if err != nil {
				return fmt.Errorf("failed to list runtimes: %w", err)
			}
			if len(idArr) == 0 {
				return fmt.Errorf("runtime with alias %q not found", cmdInstanceModulesAddOptions.alias)
			}
			if len(idArr) > 1 {
				return fmt.Errorf("multiple runtimes with alias %q found: %v", cmdInstanceModulesAddOptions.alias, pie.Map(idArr, func(x e2ekeb.InstanceDetails) string {
					return x.RuntimeID
				}))
			}
			cmdInstanceModulesAddOptions.runtimeID = idArr[0].RuntimeID
		}

		clnt, err := keb.CreateInstanceClient(rootCtx, cmdInstanceModulesAddOptions.runtimeID)
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
			if m.Name == cmdInstanceModulesAddOptions.moduleName {
				isFound = true
				break
			}
		}

		if isFound {
			fmt.Println("Module is already added")
			return nil
		}

		kyma.Spec.Modules = append(kyma.Spec.Modules, operatorv1beta2.Module{
			Name: cmdInstanceModulesAddOptions.moduleName,
		})

		err = clnt.Update(rootCtx, kyma)
		if err != nil {
			return fmt.Errorf("failed to update SKR kyma: %w", err)
		}

		fmt.Println("Module is added")

		if cmdInstanceModulesAddOptions.waitDone && cmdInstanceModulesAddOptions.moduleName == "cloud-manager" {
			fmt.Printf("Waiting for module to be ready with timeout %s\n", cmdInstanceModulesAddOptions.timeout.String())

			logger := logr.Discard()
			if verbose {
				logger = rootLogger.WithName("waitModuleReady")
			}
			err := wait.PollUntilContextTimeout(rootCtx, 5*time.Second, cmdInstanceModulesAddOptions.timeout, false, func(ctx context.Context) (done bool, err error) {
				err = clnt.Get(rootCtx, types.NamespacedName{
					Namespace: "kyma-system",
					Name:      "default",
				}, kyma)
				if err != nil {
					return false, fmt.Errorf("failed to get SKR kyma while waiting module ready: %w", err)
				}
				m, ok := kyma.GetModuleStatusMap()[cmdInstanceModulesAddOptions.moduleName]
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
	cmdInstanceModulesAdd.Flags().StringVarP(&cmdInstanceModulesAddOptions.runtimeID, "runtime-id", "r", "", "The runtime ID")
	cmdInstanceModulesAdd.Flags().StringVarP(&cmdInstanceModulesAddOptions.alias, "alias", "a", "", "The runtime alias")
	cmdInstanceModulesAdd.Flags().StringVarP(&cmdInstanceModulesAddOptions.moduleName, "module", "m", "", "The module name")
	cmdInstanceModulesAdd.Flags().BoolVarP(&cmdInstanceModulesAddOptions.waitDone, "wait", "w", false, "Wait until module is added")
	cmdInstanceModulesAdd.Flags().DurationVarP(&cmdInstanceModulesAddOptions.timeout, "timeout", "t", 5*time.Minute, "Timeout")
	_ = cmdInstanceModulesAdd.MarkFlagRequired("module")
	cmdInstanceModulesAdd.MarkFlagsMutuallyExclusive("runtime-id", "alias")
	cmdInstanceModulesAdd.MarkFlagsOneRequired("runtime-id", "alias")
}
