package main

import (
	"fmt"
	"time"

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
	cmdInstanceDelete.Flags().StringVarP(&runtimeID, "runtime-id", "r", "", "Alias name for the instance")
	cmdInstanceDelete.Flags().BoolVarP(&waitDone, "wait", "w", false, "Wait for instance to be ready before exiting")
	cmdInstanceDelete.Flags().DurationVarP(&timeout, "timeout", "t", 900*time.Second, "Timeout for waiting for instance to become ready")

	_ = cmdInstanceDelete.MarkFlagRequired("alias")
	_ = cmdInstanceDelete.MarkFlagRequired("provider")

	cmdInstance.AddCommand(cmdInstanceDelete)
}
