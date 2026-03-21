package main

import (
	"fmt"
	"time"

	e2ekeb "github.com/kyma-project/cloud-manager/e2e/keb"
	"github.com/spf13/cobra"
)

var cmdInstanceWait = &cobra.Command{
	Use: "wait",
	RunE: func(cmd *cobra.Command, args []string) error {
		keb, err := e2ekeb.Create(rootCtx, config)
		if err != nil {
			return fmt.Errorf("failed to create keb: %w", err)
		}

		var opts []e2ekeb.WaitOption
		if runtimeID != "" {
			opts = append(opts, e2ekeb.WithRuntime(runtimeID))
		}
		if alias != "" {
			opts = append(opts, e2ekeb.WithAlias(alias))
		}
		if timeout > 0 {
			opts = append(opts, e2ekeb.WithTimeout(timeout))
		}
		if verbose {
			opts = append(opts, e2ekeb.WaitProgressPrint())
		}

		err = e2ekeb.WaitCompleted(rootCtx, keb, opts...)
		return err
	},
}

func init() {
	cmdInstanceWait.Flags().StringVarP(&alias, "alias", "a", "", "alias of instance to wait for")
	cmdInstanceWait.Flags().StringVarP(&runtimeID, "runtime", "r", "", "runtime ID of instance to wait for")
	cmdInstanceWait.Flags().BoolVarP(&all, "all", "", false, "wait for all runtime instances")
	cmdInstanceWait.Flags().DurationVarP(&timeout, "timeout", "t", 900*time.Second, "timeout for waiting for instance to become ready")
	cmdInstanceWait.MarkFlagsOneRequired("runtime", "alias", "all")

	cmdInstance.AddCommand(cmdInstanceWait)
}
