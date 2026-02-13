package main

import (
	"fmt"

	"github.com/elliotchance/pie/v2"
	"github.com/kyma-project/cloud-manager/e2e"
	e2ekeb "github.com/kyma-project/cloud-manager/e2e/keb"
	"github.com/spf13/cobra"
)

var (
	cleanVerbose bool
)

var cmdInstanceClean = &cobra.Command{
	Use:   "clean",
	Short: "Clean orphaned cloud resources from an SKR instance after e2e test failures",
	Long: `Delete all SKR cloud-manager resources from the specified instance.

This command is useful for cleaning up orphaned resources that remain after e2e test failures.
When tests complete successfully, they typically clean up after themselves.
However, failed tests may leave resources undeleted.

This cleanup runs using DeleteAllOf for efficiency and should be executed after tests complete
but before deleting the instance itself.

Examples:
  # Clean instance by runtime ID
  e2e instance clean --runtime-id abc-123

  # Clean instance by alias
  e2e instance clean --alias my-test-cluster

  # Clean with verbose logging
  e2e instance clean --runtime-id abc-123 --verbose
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		keb, err := e2ekeb.Create(rootCtx, config)
		if err != nil {
			return fmt.Errorf("failed to create keb: %w", err)
		}

		if runtimeID == "" {
			idArr, err := keb.List(rootCtx, e2ekeb.WithAlias(alias))
			if err != nil {
				return fmt.Errorf("failed to list runtimes: %w", err)
			}
			if len(idArr) == 0 {
				return fmt.Errorf("runtime with alias %q not found", alias)
			}
			if len(idArr) > 1 {
				return fmt.Errorf("multiple runtimes with alias %q found: %v", alias, pie.Map(idArr, func(x e2ekeb.InstanceDetails) string {
					return x.RuntimeID
				}))
			}
			runtimeID = idArr[0].RuntimeID
		}

		if cleanVerbose {
			fmt.Printf("Cleaning SKR instance %s...\n", runtimeID)
		}

		instance, err := keb.GetInstance(rootCtx, runtimeID)
		if err != nil {
			return fmt.Errorf("failed to get instance details: %w", err)
		}

		if cleanVerbose {
			fmt.Printf("Instance: %s (provider: %s, shoot: %s)\n", instance.Alias, instance.Provider, instance.ShootName)
			fmt.Println("Connecting to SKR instance...")
		}

		skrClient, err := keb.CreateInstanceClient(rootCtx, runtimeID)
		if err != nil {
			return fmt.Errorf("failed to create SKR client: %w", err)
		}

		if cleanVerbose {
			fmt.Println("Starting cleanup of cloud resources...")
		}

		err = e2e.CleanSkrNoWait(rootCtx, skrClient)
		if err != nil {
			return fmt.Errorf("failed to clean SKR: %w", err)
		}

		fmt.Printf("\n✅ Successfully cleaned SKR instance %s\n", runtimeID)

		return nil
	},
}

func init() {
	cmdInstance.AddCommand(cmdInstanceClean)
	cmdInstanceClean.Flags().StringVarP(&runtimeID, "runtime-id", "r", "", "Runtime ID of the instance to clean")
	cmdInstanceClean.Flags().StringVarP(&alias, "alias", "a", "", "Alias of the instance to clean")
	cmdInstanceClean.Flags().BoolVarP(&cleanVerbose, "verbose", "v", false, "Enable verbose logging")
	cmdInstanceClean.MarkFlagsMutuallyExclusive("runtime-id", "alias")
	cmdInstanceClean.MarkFlagsOneRequired("runtime-id", "alias")
}
