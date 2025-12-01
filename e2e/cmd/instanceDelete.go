package main

import (
	"fmt"
	"time"

	"github.com/elliotchance/pie/v2"
	e2ekeb "github.com/kyma-project/cloud-manager/e2e/keb"
	"github.com/spf13/cobra"
)

var cmdInstanceDelete = &cobra.Command{
	Use:   "delete",
	Short: "Delete an instance with given runtime id, and optionally wait until it's deleted.",
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

		err = keb.DeleteInstance(
			rootCtx,
			e2ekeb.WithRuntime(runtimeID),
			e2ekeb.WithTimeout(timeout),
		)
		if err != nil {
			return fmt.Errorf("failed to delete instance: %w", err)
		}

		fmt.Println("Instance is marked for deletion.")

		if waitDone {
			fmt.Printf("Waiting for instance to be destroyed...")
			err = keb.WaitDeleted(rootCtx, e2ekeb.WithRuntime(runtimeID), e2ekeb.WithTimeout(timeout))
			if err != nil {
				return fmt.Errorf("failed to wait for instance to be deleted: %w", err)
			}
			fmt.Println("Instance is destroyed.")
		}

		return nil
	},
}

func init() {
	cmdInstance.AddCommand(cmdInstanceDelete)
	cmdInstanceDelete.Flags().StringVarP(&runtimeID, "runtime-id", "r", "", "Alias name for the instance")
	cmdInstanceDelete.Flags().StringVarP(&alias, "alias", "a", "", "The runtime alias")
	cmdInstanceDelete.Flags().BoolVarP(&waitDone, "wait", "w", false, "Wait for instance to be ready before exiting")
	cmdInstanceDelete.Flags().DurationVarP(&timeout, "timeout", "t", 900*time.Second, "Timeout for waiting for instance to become ready")
	cmdInstanceDelete.MarkFlagsMutuallyExclusive("runtime-id", "alias")
	cmdInstanceDelete.MarkFlagsOneRequired("runtime-id", "alias")
}
