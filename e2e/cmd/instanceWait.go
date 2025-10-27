package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/elliotchance/pie/v2"
	"github.com/kyma-project/cloud-manager/e2e"
	"github.com/kyma-project/cloud-manager/e2e/sim"
	"github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"
	"github.com/spf13/cobra"
)

var all bool
var runtimes []string
var timeoutSeconds int

var cmdInstanceWait = &cobra.Command{
	Use: "wait",
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := rootLogger.WithName("instance-wait")

		f := e2e.NewWorldFactory()
		world, err := f.Create(rootCtx, e2e.WorldCreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create world: %w", err)
		}

		if all {
			arr, err := world.Sim().Keb().List(rootCtx)
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
			logger.Info("Nothing to wait for")
			return nil
		}

		waitOpts := []sim.WaitOption{
			sim.WithTimeout(time.Duration(timeoutSeconds) * time.Second),
			sim.WithProgressCallback(func(p sim.WaitProgress) {
				logger.
					WithValues(
						"done", strings.Join(p.DoneAliases(), " "),
						"pending", strings.Join(p.PendingAliases(), " "),
					).
					Info("Wait progress")
			}),
		}
		if len(runtimes) > 0 {
			waitOpts = append(waitOpts, sim.WithRuntimes(runtimes))
		}

		err = world.Sim().Keb().WaitProvisioningCompleted(rootCtx, waitOpts...)
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
