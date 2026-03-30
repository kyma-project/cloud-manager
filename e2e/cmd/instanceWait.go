package main

import (
	"fmt"
	"time"

	e2ekeb "github.com/kyma-project/cloud-manager/e2e/keb"
	"github.com/spf13/cobra"
)

type cmdInstanceWaitOptionsType struct {
	alias     string
	runtimeID string
	all       bool
	timeout   time.Duration
}

var cmdInstanceWaitOptions cmdInstanceWaitOptionsType

var cmdInstanceWait = &cobra.Command{
	Use: "wait",
	RunE: func(cmd *cobra.Command, args []string) error {
		keb, err := e2ekeb.Create(rootCtx, config)
		if err != nil {
			return fmt.Errorf("failed to create keb: %w", err)
		}

		var opts []e2ekeb.WaitOption
		if cmdInstanceWaitOptions.runtimeID != "" {
			opts = append(opts, e2ekeb.WithRuntime(cmdInstanceWaitOptions.runtimeID))
		}
		if cmdInstanceWaitOptions.alias != "" {
			opts = append(opts, e2ekeb.WithAlias(cmdInstanceWaitOptions.alias))
		}
		if cmdInstanceWaitOptions.timeout > 0 {
			opts = append(opts, e2ekeb.WithTimeout(cmdInstanceWaitOptions.timeout))
		}
		if verbose {
			opts = append(opts, e2ekeb.WaitProgressPrint(), e2ekeb.WithLogger(rootLogger))
		}

		err = e2ekeb.WaitCompleted(rootCtx, keb, opts...)
		return err
	},
}

func init() {
	cmdInstanceWait.Flags().StringVarP(&cmdInstanceWaitOptions.alias, "alias", "a", "", "alias of instance to wait for")
	cmdInstanceWait.Flags().StringVarP(&cmdInstanceWaitOptions.runtimeID, "runtime", "r", "", "runtime ID of instance to wait for")
	cmdInstanceWait.Flags().BoolVarP(&cmdInstanceWaitOptions.all, "all", "", false, "wait for all runtime instances")
	cmdInstanceWait.Flags().DurationVarP(&cmdInstanceWaitOptions.timeout, "timeout", "t", 900*time.Second, "timeout for waiting for instance to become ready")
	cmdInstanceWait.MarkFlagsOneRequired("runtime", "alias", "all")

	cmdInstance.AddCommand(cmdInstanceWait)
}
