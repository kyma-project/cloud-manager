package main

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/e2e"
	"github.com/kyma-project/cloud-manager/e2e/sim"
	"github.com/spf13/cobra"
)

var alias string
var provider string
var waitDone bool

var cmdInstanceCreate = &cobra.Command{
	Use:   "create",
	Short: "Create an instance with given alias and provider, and optionally wait until it is provisioned",
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := rootLogger.WithName("instance-create")
		pt, err := cloudcontrolv1beta1.ParseProviderType(provider)
		if err != nil {
			return err
		}

		f := e2e.NewWorldFactory()
		world, err := f.Create(rootCtx, e2e.WorldCreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create world: %w", err)
		}

		id, err := world.Sim().Keb().CreateInstance(rootCtx,
			sim.WithAlias(alias),
			sim.WithGlobalAccount(uuid.NewString()),
			sim.WithSubAccount(uuid.NewString()),
			sim.WithProvider(pt),
		)
		if err != nil {
			return fmt.Errorf("error creating instance: %w", err)
		}

		logger = id.AddLoggerValues(logger)
		logger.Info("Instance created")

		if waitDone {
			logger.Info("Waiting for instance to be ready")
			err = world.Sim().Keb().WaitProvisioningCompleted(rootCtx, sim.WithRuntime(id.RuntimeID), sim.WithTimeout(time.Duration(timeoutSeconds)*time.Second))
			if err != nil {
				return fmt.Errorf("error waiting provisioning completed: %w", err)
			}
			logger.Info("Instance is ready")
		}

		return nil
	},
}

func init() {
	cmdInstanceCreate.Flags().StringVarP(&alias, "alias", "a", "", "Alias name for the instance")
	cmdInstanceCreate.Flags().StringVarP(&provider, "provider", "p", "", "Provider name for the instance")
	cmdInstanceCreate.Flags().BoolVarP(&waitDone, "wait", "w", false, "Wait for instance to be ready before exiting")
	cmdInstanceCreate.Flags().IntVarP(&timeoutSeconds, "timeout", "t", 900, "Timeout in seconds for waiting for instance to become ready")

	_ = cmdInstanceCreate.MarkFlagRequired("alias")
	_ = cmdInstanceCreate.MarkFlagRequired("provider")

	cmdInstance.AddCommand(cmdInstanceCreate)
}
