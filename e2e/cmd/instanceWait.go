package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/elliotchance/pie/v2"
	e2ekeb "github.com/kyma-project/cloud-manager/e2e/keb"
	"github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"
	"github.com/spf13/cobra"
)

var all bool
var runtimes []string
var timeoutSeconds int

var cmdInstanceWait = &cobra.Command{
	Use: "wait",
	RunE: func(cmd *cobra.Command, args []string) error {
		keb, err := e2ekeb.Create(rootCtx, config)
		if err != nil {
			return fmt.Errorf("failed to create keb: %w", err)
		}

		if all {
			arr, err := keb.List(rootCtx)
			if err != nil {
				return fmt.Errorf("failed to list instances for wait them all: %w", err)
			}
			for _, id := range arr {
				if id.BeingDeleted || id.State == infrastructuremanagerv1.RuntimeStateFailed {
					continue
				}
				runtimes = append(runtimes, id.RuntimeID)
			}
			runtimes = pie.Unique(runtimes)
		}

		if len(runtimes) == 0 {
			fmt.Println("No runtimes found")
			return nil
		}

		waitOpts := []e2ekeb.WaitOption{
			e2ekeb.WithTimeout(time.Duration(timeoutSeconds) * time.Second),
			e2ekeb.WithProgressCallback(func(p e2ekeb.WaitProgress) {
				fmt.Printf("Done: %s\n", strings.Join(p.DoneAliases(), " "))
				fmt.Printf("Pending: %s\n", strings.Join(p.PendingAliases(), " "))
			}),
		}
		if len(runtimes) > 0 {
			waitOpts = append(waitOpts, e2ekeb.WithRuntimes(runtimes))
		}

		err = keb.WaitProvisioningCompleted(rootCtx, waitOpts...)
		if err != nil {
			return fmt.Errorf("error waiting for instance(s) to become ready: %w", err)
		}

		return nil
	},
}

func init() {
	cmdInstanceWait.Flags().StringSliceVarP(&runtimes, "runtime", "r", []string{}, "runtime ID to wait for")
	cmdInstanceWait.Flags().BoolVarP(&all, "all", "a", false, "wait for all runtimes")
	cmdInstanceWait.Flags().IntVarP(&timeoutSeconds, "timeout", "t", 900, "Timeout in seconds for waiting for instance to become ready")
	cmdInstanceWait.MarkFlagsOneRequired("runtime", "all")

	cmdInstance.AddCommand(cmdInstanceWait)
}
