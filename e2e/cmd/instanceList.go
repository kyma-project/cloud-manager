package main

import (
	"fmt"

	"github.com/kyma-project/cloud-manager/e2e"
	"github.com/spf13/cobra"
)

var cmdInstanceList = &cobra.Command{
	Use:   "list",
	Short: "List all instances",
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := rootLogger.WithName("instance-list")

		f := e2e.NewWorldFactory()
		world, err := f.Create(rootCtx, e2e.WorldCreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create world: %w", err)
		}

		arr, err := world.Sim().Keb().List(rootCtx)
		if err != nil {
			return fmt.Errorf("error listing intances: %w", err)
		}

		for _, id := range arr {
			logger := id.AddLoggerValues(logger)
			logger.Info(fmt.Sprintf("Instance %s", id.Alias))
		}

		if len(arr) == 0 {
			logger.Info("No instances found")
		}

		return nil
	},
}

func init() {
	cmdInstance.AddCommand(cmdInstanceList)
}
