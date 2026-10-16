package main

import (
	"fmt"

	"github.com/google/uuid"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	e2ekeb "github.com/kyma-project/cloud-manager/e2e/keb"
	e2elib "github.com/kyma-project/cloud-manager/e2e/lib"
	e2esim "github.com/kyma-project/cloud-manager/e2e/sim"
	"github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

type cmdManifestDebugOptionsType struct {
	alias    string
	provider string
	shoot    bool
	runtime  bool
}

var cmdManifestDebugOptions cmdManifestDebugOptionsType

var cmdManifestDebug = &cobra.Command{
	Use: "debug",
	RunE: func(cmd *cobra.Command, args []string) error {
		pt, err := cloudcontrolv1beta1.ParseProviderType(cmdManifestDebugOptions.provider)
		if err != nil {
			return err
		}
		subscription := config.Subscriptions.GetDefaultForProvider(pt)
		if subscription == nil {
			return fmt.Errorf("no subscription found for provider %s", pt)
		}

		keb, err := e2ekeb.Create(rootCtx, config)
		if err != nil {
			return fmt.Errorf("failed to create keb: %w", err)
		}

		cpr, err := keb.LoadCloudProfileRegistry(rootCtx)
		if err != nil {
			return fmt.Errorf("failed to load cloud profile registry: %w", err)
		}

		rtBuilder := e2elib.NewRuntimeBuilder(cpr, config).
			WithAlias(cmdManifestDebugOptions.alias).
			WithProvider(pt, e2elib.DefaultRegions[pt]).
			WithSecretBindingName(subscription.Name).
			WithGlobalAccount(uuid.NewString()).
			WithSubAccount(uuid.NewString())
		if err := rtBuilder.Validate(); err != nil {
			return fmt.Errorf("invalid create instance request: %w", err)
		}

		rt := rtBuilder.Build()
		rt.Finalizers = append(rt.Finalizers, infrastructuremanagerv1.Finalizer)

		if cmdManifestDebugOptions.runtime {
			rtyaml, err := yaml.Marshal(rt)
			if err != nil {
				return fmt.Errorf("failed to marshal runtime to yaml: %w", err)
			}
			fmt.Println(string(rtyaml))
		}

		if cmdManifestDebugOptions.shoot {
			if cmdManifestDebugOptions.runtime {
				fmt.Println("---")
				fmt.Println()
			}

			shootBuilder := e2esim.NewShootBuilder(cpr, config).
				WithRuntime(rt)
			if err := shootBuilder.Validate(); err != nil {
				return fmt.Errorf("invalid shoot request: %w", err)
			}

			shoot := shootBuilder.Build()
			shootyaml, err := yaml.Marshal(shoot)
			if err != nil {
				return fmt.Errorf("failed to marshal shoot to yaml: %w", err)
			}
			fmt.Println(string(shootyaml))
		}

		return nil
	},
}

func init() {
	cmdManifestDebug.Flags().StringVarP(&cmdManifestDebugOptions.alias, "alias", "a", "", "Alias name for the instance")
	cmdManifestDebug.Flags().StringVarP(&cmdManifestDebugOptions.provider, "provider", "p", "", "Provider name for the instance")
	cmdManifestDebug.Flags().BoolVar(&cmdManifestDebugOptions.shoot, "shoot", true, "Print the Shoot manifest")
	cmdManifestDebug.Flags().BoolVar(&cmdManifestDebugOptions.runtime, "runtime", false, "Print the Runtime manifest")

	_ = cmdInstanceCreate.MarkFlagRequired("alias")
	_ = cmdInstanceCreate.MarkFlagRequired("provider")

	cmdManifest.AddCommand(cmdManifestDebug)
}
