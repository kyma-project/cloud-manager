package main

import (
	"fmt"
	"time"

	"github.com/elliotchance/pie/v2"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	e2eclean "github.com/kyma-project/cloud-manager/e2e/clean"
	e2ekeb "github.com/kyma-project/cloud-manager/e2e/keb"
	commonscheme "github.com/kyma-project/cloud-manager/pkg/common/scheme"
	"github.com/spf13/cobra"
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

		skrClient, err := keb.CreateInstanceClient(rootCtx, runtimeID)
		if err != nil {
			return fmt.Errorf("failed to create SKR client: %w", err)
		}

		opts := []e2eclean.Option{
			e2eclean.WithClient(skrClient),
			e2eclean.WithScheme(commonscheme.SkrScheme),
			e2eclean.WithMatchers(
				e2eclean.MatchAll(
					e2eclean.MatchingGroup(cloudresourcesv1beta1.GroupVersion.Group),
					e2eclean.NotMatch(e2eclean.MatchingKind("CloudResources")),
				),
			),
			e2eclean.WithTimeout(timeout),
			e2eclean.WithWait(waitDone),
			e2eclean.WithForceDeleteOnTimeout(force),
			e2eclean.WithDryRun(dryRun),
		}
		if verbose {
			opts = append(opts, e2eclean.WithLogger(rootLogger))
		}

		err = e2eclean.Clean(rootCtx, opts...)
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
	cmdInstanceClean.Flags().BoolVarP(&waitDone, "wait", "w", false, "Wait until deleted")
	cmdInstanceClean.Flags().DurationVarP(&timeout, "timeout", "t", 30*time.Minute, "Timeout")
	cmdInstanceClean.Flags().BoolVarP(&force, "force", "f", false, "Force delete after timeout")
	cmdInstanceClean.Flags().BoolVarP(&dryRun, "dry-run", "", false, "Dry run")
	cmdInstanceClean.MarkFlagsMutuallyExclusive("runtime-id", "alias")
	cmdInstanceClean.MarkFlagsOneRequired("runtime-id", "alias")
}
