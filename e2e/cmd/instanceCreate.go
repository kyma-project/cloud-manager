package main

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	e2ekeb "github.com/kyma-project/cloud-manager/e2e/keb"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type cmdInstanceCreateOptionsType struct {
	alias               string
	provider            string
	waitDone            bool
	timeout             time.Duration
	shootCreatedTimeout time.Duration
	reuse               bool
}

var cmdInstanceCreateOptions cmdInstanceCreateOptionsType

var cmdInstanceCreate = &cobra.Command{
	Use:   "create",
	Short: "Create an instance with given alias and provider, and optionally wait until it is provisioned",
	RunE: func(cmd *cobra.Command, args []string) error {
		pt, err := cloudcontrolv1beta1.ParseProviderType(cmdInstanceCreateOptions.provider)
		if err != nil {
			return err
		}

		keb, err := e2ekeb.Create(rootCtx, config)
		if err != nil {
			return fmt.Errorf("failed to create keb: %w", err)
		}

		var id *e2ekeb.InstanceDetails

		if cmdInstanceCreateOptions.reuse {
			existing, err := keb.List(rootCtx, e2ekeb.WithAlias(cmdInstanceCreateOptions.alias))
			if err != nil {
				return fmt.Errorf("error listing instances: %w", err)
			}
			if len(existing) > 0 {
				fmt.Printf("Reusing existing instance with alias %q (runtimeID: %s)\n", cmdInstanceCreateOptions.alias, existing[0].RuntimeID)
				id = &existing[0]
			}
		}

		if id == nil {
			created, err := keb.CreateInstance(rootCtx,
				e2ekeb.WithAlias(cmdInstanceCreateOptions.alias),
				e2ekeb.WithGlobalAccount(uuid.NewString()),
				e2ekeb.WithSubAccount(uuid.NewString()),
				e2ekeb.WithProvider(pt),
				e2ekeb.WithTimeout(cmdInstanceCreateOptions.shootCreatedTimeout),
			)
			if err != nil {
				return fmt.Errorf("error creating instance: %w", err)
			}
			id = &created
		}

		b, err := yaml.Marshal(id)
		if err != nil {
			return fmt.Errorf("error marshalling instance details to yaml: %w", err)
		}
		fmt.Println("Instance created:")
		fmt.Println(string(b))

		if cmdInstanceCreateOptions.waitDone {
			fmt.Printf("Waiting for instance to be ready with timeout of %s...\n", cmdInstanceCreateOptions.timeout)
			opts := []e2ekeb.WaitOption{e2ekeb.WithAlias(id.Alias), e2ekeb.WithTimeout(cmdInstanceCreateOptions.timeout), e2ekeb.WithErrorDuration(cmdInstanceCreateOptions.timeout), e2ekeb.WithTerminalErrorDuration(5 * time.Minute)}
			if verbose {
				opts = append(opts, e2ekeb.WaitProgressPrint())
			}
			err = e2ekeb.WaitCompleted(rootCtx, keb, opts...)
			if err != nil {
				return fmt.Errorf("error waiting provisioning completed: %w", err)
			}
			fmt.Println("Instance is ready")
		}

		return nil
	},
}

func init() {
	cmdInstanceCreate.Flags().StringVarP(&cmdInstanceCreateOptions.alias, "alias", "a", "", "Alias name for the instance")
	cmdInstanceCreate.Flags().StringVarP(&cmdInstanceCreateOptions.provider, "provider", "p", "", "Provider name for the instance")
	cmdInstanceCreate.Flags().BoolVarP(&cmdInstanceCreateOptions.waitDone, "wait", "w", false, "Wait for instance to be ready before exiting")
	cmdInstanceCreate.Flags().DurationVarP(&cmdInstanceCreateOptions.timeout, "timeout", "t", 900*time.Second, "Timeout for waiting for instance to become ready")
	cmdInstanceCreate.Flags().DurationVarP(&cmdInstanceCreateOptions.shootCreatedTimeout, "shoot-created-timeout", "s", 60*time.Second, "Timeout for waiting for the shoot object to appear in Garden after create")
	cmdInstanceCreate.Flags().BoolVar(&cmdInstanceCreateOptions.reuse, "reuse", false, "Reuse existing instance with the same alias instead of creating a new one")

	_ = cmdInstanceCreate.MarkFlagRequired("alias")
	_ = cmdInstanceCreate.MarkFlagRequired("provider")

	cmdInstance.AddCommand(cmdInstanceCreate)
}
