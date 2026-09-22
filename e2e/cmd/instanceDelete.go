package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/elliotchance/pie/v2"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	e2ekeb "github.com/kyma-project/cloud-manager/e2e/keb"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"
)

type cmdInstanceDeleteOptionsType struct {
	runtimeID string
	alias     string
	waitDone  bool
	timeout   time.Duration
}

var cmdInstanceDeleteOptions cmdInstanceDeleteOptionsType

var cmdInstanceDelete = &cobra.Command{
	Use:   "delete",
	Short: "Delete an instance with given runtime id, and optionally wait until it's deleted.",
	RunE: func(cmd *cobra.Command, args []string) error {
		keb, err := e2ekeb.Create(rootCtx, config)
		if err != nil {
			return fmt.Errorf("failed to create keb: %w", err)
		}

		if cmdInstanceDeleteOptions.runtimeID == "" {
			idArr, err := keb.List(rootCtx, e2ekeb.WithAlias(cmdInstanceDeleteOptions.alias))
			if err != nil {
				return fmt.Errorf("failed to list runtimes: %w", err)
			}
			if len(idArr) == 0 {
				return fmt.Errorf("runtime with alias %q not found", cmdInstanceDeleteOptions.alias)
			}
			if len(idArr) > 1 {
				return fmt.Errorf("multiple runtimes with alias %q found: %v", cmdInstanceDeleteOptions.alias, pie.Map(idArr, func(x e2ekeb.InstanceDetails) string {
					return x.RuntimeID
				}))
			}
			cmdInstanceDeleteOptions.runtimeID = idArr[0].RuntimeID
		}

		err = keb.DeleteInstance(
			rootCtx,
			e2ekeb.WithRuntime(cmdInstanceDeleteOptions.runtimeID),
			e2ekeb.WithTimeout(cmdInstanceDeleteOptions.timeout),
		)
		if err != nil {
			return fmt.Errorf("failed to delete instance: %w", err)
		}

		fmt.Println("Instance is marked for deletion.")

		if cmdInstanceDeleteOptions.waitDone {
			fmt.Printf("Waiting for instance to be destroyed...")
			opts := []e2ekeb.WaitOption{
				e2ekeb.WithRuntime(cmdInstanceDeleteOptions.runtimeID),
				e2ekeb.WithTimeout(cmdInstanceDeleteOptions.timeout),
			}
			if verbose {
				opts = append(opts, e2ekeb.WithLogger(rootLogger), e2ekeb.WaitProgressPrint())
			}
			err = e2ekeb.WaitCompleted(rootCtx, keb, opts...)
			if err != nil {
				printShootStatus(keb, cmdInstanceDeleteOptions.runtimeID)
				return fmt.Errorf("failed to wait for instance to be deleted: %w", err)
			}
			fmt.Println("Instance is destroyed.")
		}

		return nil
	},
}

func init() {
	cmdInstance.AddCommand(cmdInstanceDelete)
	cmdInstanceDelete.Flags().StringVarP(&cmdInstanceDeleteOptions.runtimeID, "runtime-id", "r", "", "Alias name for the instance")
	cmdInstanceDelete.Flags().StringVarP(&cmdInstanceDeleteOptions.alias, "alias", "a", "", "The runtime alias")
	cmdInstanceDelete.Flags().BoolVarP(&cmdInstanceDeleteOptions.waitDone, "wait", "w", false, "Wait for instance to be deleted before exiting")
	cmdInstanceDelete.Flags().DurationVarP(&cmdInstanceDeleteOptions.timeout, "timeout", "t", 40*time.Minute, "Timeout for waiting for instance to be deleted")
	cmdInstanceDelete.MarkFlagsMutuallyExclusive("runtime-id", "alias")
	cmdInstanceDelete.MarkFlagsOneRequired("runtime-id", "alias")
}

func printShootStatus(keb e2ekeb.Keb, runtimeID string) {
	instances, err := keb.List(rootCtx, e2ekeb.WithRuntime(runtimeID))
	if err != nil || len(instances) == 0 {
		fmt.Printf("shoot status: runtime %s not found in KEB\n", runtimeID)
		return
	}
	shootName := instances[0].ShootName
	if shootName == "" {
		fmt.Printf("shoot status: no shoot name for runtime %s\n", runtimeID)
		return
	}
	shoot := &gardencorev1beta1.Shoot{}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = keb.GardenClient().Get(ctx, types.NamespacedName{
		Namespace: keb.Config().GardenNamespace,
		Name:      shootName,
	}, shoot)
	if err != nil {
		fmt.Printf("shoot status: failed to get shoot %s: %v\n", shootName, err)
		return
	}
	b, err := json.MarshalIndent(shoot.Status, "", "  ")
	if err != nil {
		fmt.Printf("shoot status: failed to marshal shoot status: %v\n", err)
		return
	}
	fmt.Printf("shoot %s status:\n%s\n", shootName, string(b))
}
