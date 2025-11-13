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

var alias string
var provider string
var waitDone bool

var cmdInstanceCreate = &cobra.Command{
	Use:   "create",
	Short: "Create an instance with given alias and provider, and optionally wait until it is provisioned",
	RunE: func(cmd *cobra.Command, args []string) error {
		pt, err := cloudcontrolv1beta1.ParseProviderType(provider)
		if err != nil {
			return err
		}

		keb, err := e2ekeb.Create(rootCtx, config)
		if err != nil {
			return fmt.Errorf("failed to create keb: %w", err)
		}

		id, err := keb.CreateInstance(rootCtx,
			e2ekeb.WithAlias(alias),
			e2ekeb.WithGlobalAccount(uuid.NewString()),
			e2ekeb.WithSubAccount(uuid.NewString()),
			e2ekeb.WithProvider(pt),
		)
		if err != nil {
			return fmt.Errorf("error creating instance: %w", err)
		}

		b, err := yaml.Marshal(id)
		if err != nil {
			return fmt.Errorf("error marshalling instance details to yaml: %w", err)
		}
		fmt.Println("Instance created:")
		fmt.Println(string(b))

		if waitDone {
			fmt.Printf("Waiting for instance to be ready with timeout of %d seconds...\n", timeoutSeconds)
			err = keb.WaitProvisioningCompleted(rootCtx, e2ekeb.WithRuntime(id.RuntimeID), e2ekeb.WithTimeout(time.Duration(timeoutSeconds)*time.Second))
			if err != nil {
				return fmt.Errorf("error waiting provisioning completed: %w", err)
			}
			fmt.Println("Instance is ready")
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
