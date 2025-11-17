package main

import (
	"context"
	"fmt"
	"time"

	"github.com/kyma-project/cloud-manager/e2e/sim"
	"github.com/spf13/cobra"
)

var cmdSimRun = &cobra.Command{
	Use: "run",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := rootCtx
		cancel := func() {}
		if timeout > 0 {
			ctx, cancel = context.WithTimeout(ctx, timeout)
		}
		defer cancel()

		fmt.Println("Starting Sim...")
		fmt.Println("")

		err := sim.Run(ctx, config)

		fmt.Println("")
		fmt.Println("Sim has stopped")
		fmt.Println("")

		return err
	},
}

func init() {
	cmdSim.AddCommand(cmdSimRun)
	cmdSimRun.Flags().DurationVarP(&timeout, "timeout", "t", time.Duration(0), "Time to run and then quit")
}
