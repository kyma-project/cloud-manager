package main

import (
	"fmt"
	"time"

	"github.com/elliotchance/pie/v2"
	"github.com/google/uuid"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	e2ekeb "github.com/kyma-project/cloud-manager/e2e/keb"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

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
			fmt.Printf("Waiting for instance to be ready with timeout of %s...\n", timeout)
			opts := []e2ekeb.WaitOption{e2ekeb.WithRuntime(id.RuntimeID), e2ekeb.WithTimeout(timeout)}
			if verbose {
				id2s := func(id e2ekeb.InstanceDetails) string {
					return fmt.Sprintf("{%s %s %s}", id.Alias, id.RuntimeID, id.ShootName)
				}
				opts = append(opts, e2ekeb.WithProgressCallback(func(progress e2ekeb.WaitProgress) {
					if progress.Changed {
						fmt.Printf("%s\n", time.Now().Format(time.RFC3339))
						fmt.Printf("Pending: %v\n", pie.Map(progress.Pending, id2s))
						fmt.Printf("WithErr: %v\n", pie.Map(progress.WithErr, id2s))
						fmt.Printf("Done: %v\n", pie.Map(progress.Done, id2s))
					}
				}))
			}
			err = keb.WaitProvisioningCompleted(rootCtx, opts...)
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
	cmdInstanceCreate.Flags().DurationVarP(&timeout, "timeout", "t", 900*time.Second, "Timeout in seconds for waiting for instance to become ready")

	_ = cmdInstanceCreate.MarkFlagRequired("alias")
	_ = cmdInstanceCreate.MarkFlagRequired("provider")

	cmdInstance.AddCommand(cmdInstanceCreate)
}
