package main

import (
	"fmt"
	"time"

	"github.com/elliotchance/pie/v2"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	e2eclean "github.com/kyma-project/cloud-manager/e2e/clean"
	e2ekeb "github.com/kyma-project/cloud-manager/e2e/keb"
	commonscheme "github.com/kyma-project/cloud-manager/pkg/common/scheme"
	"github.com/kyma-project/cloud-manager/pkg/external/operatorv1beta2"
	"github.com/spf13/cobra"
)

type cmdInstanceCleanOptionsType struct {
	runtimeID string
	alias     string
	waitDone  bool
	timeout   time.Duration
	force     bool
	dryRun    bool
	all       bool
}

var cmdInstanceCleanOptions cmdInstanceCleanOptionsType

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

		if cmdInstanceCleanOptions.runtimeID == "" {
			idArr, err := keb.List(rootCtx, e2ekeb.WithAlias(cmdInstanceCleanOptions.alias))
			if err != nil {
				return fmt.Errorf("failed to list runtimes: %w", err)
			}
			if len(idArr) == 0 {
				return fmt.Errorf("runtime with alias %q not found", cmdInstanceCleanOptions.alias)
			}
			if len(idArr) > 1 {
				return fmt.Errorf("multiple runtimes with alias %q found: %v", cmdInstanceCleanOptions.alias, pie.Map(idArr, func(x e2ekeb.InstanceDetails) string {
					return x.RuntimeID
				}))
			}
			cmdInstanceCleanOptions.runtimeID = idArr[0].RuntimeID
		}

		skrClient, err := keb.CreateInstanceClient(rootCtx, cmdInstanceCleanOptions.runtimeID)
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
			e2eclean.WithTimeout(cmdInstanceCleanOptions.timeout),
			e2eclean.WithWait(cmdInstanceCleanOptions.waitDone),
			e2eclean.WithForceDeleteOnTimeout(cmdInstanceCleanOptions.force),
			e2eclean.WithDryRun(cmdInstanceCleanOptions.dryRun),
		}
		if cmdInstanceCleanOptions.all {
			opts = append(opts, e2eclean.WithMatchers(
				e2eclean.MatchingGroup(cloudresourcesv1beta1.GroupVersion.Group),
				e2eclean.MatchingGroup(operatorv1beta2.GroupVersion.Group),
			))
		} else {
			opts = append(opts, e2eclean.WithMatchers(
				e2eclean.MatchAll(
					e2eclean.MatchingGroup(cloudresourcesv1beta1.GroupVersion.Group),
					e2eclean.NotMatch(e2eclean.MatchingKind("CloudResources")),
				),
			))

		}
		if verbose {
			opts = append(opts, e2eclean.WithLogger(rootLogger))
		}

		err = e2eclean.Clean(rootCtx, opts...)
		if err != nil {
			return fmt.Errorf("failed to clean SKR: %w", err)
		}

		fmt.Printf("\n✅ Successfully cleaned SKR instance %s\n", cmdInstanceCleanOptions.runtimeID)

		return nil
	},
}

func init() {
	cmdInstance.AddCommand(cmdInstanceClean)
	cmdInstanceClean.Flags().StringVarP(&cmdInstanceCleanOptions.runtimeID, "runtime-id", "r", "", "Runtime ID of the instance to clean")
	cmdInstanceClean.Flags().StringVarP(&cmdInstanceCleanOptions.alias, "alias", "a", "", "Alias of the instance to clean")
	cmdInstanceClean.Flags().BoolVarP(&cmdInstanceCleanOptions.waitDone, "wait", "w", false, "Wait until deleted")
	cmdInstanceClean.Flags().DurationVarP(&cmdInstanceCleanOptions.timeout, "timeout", "t", 30*time.Minute, "Timeout to wait until deleted")
	cmdInstanceClean.Flags().BoolVarP(&cmdInstanceCleanOptions.force, "force", "f", false, "Force delete after timeout")
	cmdInstanceClean.Flags().BoolVarP(&cmdInstanceCleanOptions.dryRun, "dry-run", "", false, "Dry run")
	cmdInstanceClean.Flags().BoolVarP(&cmdInstanceCleanOptions.all, "all", "", false, "Delete all from kyma groups cloud-resources and operator. Destructive! Deletes Kyma and CloudResources CR! If false it will delete all from cloud-resources except for CloudResources CR")
	cmdInstanceClean.MarkFlagsMutuallyExclusive("runtime-id", "alias")
	cmdInstanceClean.MarkFlagsOneRequired("runtime-id", "alias")
}
