package main

import (
	"fmt"

	"github.com/kyma-project/cloud-manager/e2e"
	"github.com/spf13/cobra"
)

var cmdSimRun = &cobra.Command{
	Use: "run",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := e2e.NewWorldFactory()
		world, err := f.Create(rootCtx, e2e.WorldCreateOptions{
			Config: config,
		})
		if err != nil {
			return fmt.Errorf("failed to create world: %w", err)
		}

		<-rootCtx.Done()

		fmt.Println("")
		fmt.Println("")
		fmt.Println("Waiting world to stop...")
		fmt.Println("")
		world.StopWaitGroup().Wait()
		fmt.Println("")
		fmt.Println("World has stopped")
		fmt.Println("")

		return world.RunError()
	},
}

func init() {
	cmdSim.AddCommand(cmdSimRun)
}
