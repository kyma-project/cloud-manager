package main

import (
	"fmt"
	"time"

	"github.com/elliotchance/pie/v2"
	e2ekeb "github.com/kyma-project/cloud-manager/e2e/keb"
	"github.com/spf13/cobra"
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