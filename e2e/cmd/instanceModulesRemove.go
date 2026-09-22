package main

import (
	"context"
	"fmt"
	"time"

	"github.com/elliotchance/pie/v2"
	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/api"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	e2ekeb "github.com/kyma-project/cloud-manager/e2e/keb"
	"github.com/kyma-project/cloud-manager/pkg/external/operatorv1beta2"
	"github.com/spf13/cobra"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type cmdInstanceModulesRemoveOptionsType struct {
	runtimeID  string
	alias      string
	moduleName string
	waitDone   bool
	timeout    time.Duration
}

var cmdInstanceModulesRemoveOptions cmdInstanceModulesRemoveOptionsType

var cmdInstanceModulesRemove = &cobra.Command{
	Use: "remove",
	RunE: func(cmd *cobra.Command, args []string) error {
		keb, err := e2ekeb.Create(rootCtx, config)
		if err != nil {
			return fmt.Errorf("failed to create keb: %w", err)
		}

		if cmdInstanceModulesRemoveOptions.runtimeID == "" {
			idArr, err := keb.List(rootCtx, e2ekeb.WithAlias(cmdInstanceModulesRemoveOptions.alias))
			if err != nil {
				return fmt.Errorf("failed to list runtimes: %w", err)
			}
			if len(idArr) == 0 {
				return fmt.Errorf("runtime with alias %q not found", cmdInstanceModulesRemoveOptions.alias)
			}
			if len(idArr) > 1 {
				return fmt.Errorf("multiple runtimes with alias %q found: %v", cmdInstanceModulesRemoveOptions.alias, pie.Map(idArr, func(x e2ekeb.InstanceDetails) string {
					return x.RuntimeID
				}))
			}
			cmdInstanceModulesRemoveOptions.runtimeID = idArr[0].RuntimeID
		}

		clnt, err := keb.CreateInstanceClient(rootCtx, cmdInstanceModulesRemoveOptions.runtimeID)
		if err != nil {
			return err
		}

		kyma := &operatorv1beta2.Kyma{}
		err = clnt.Get(rootCtx, types.NamespacedName{
			Namespace: "kyma-system",
			Name:      "default",
		}, kyma)
		if err != nil {
			return fmt.Errorf("failed to get SKR kyma: %w", err)
		}

		isFound := false
		for _, m := range kyma.Spec.Modules {
			if m.Name == cmdInstanceModulesRemoveOptions.moduleName {
				isFound = true
				break
			}
		}

		if cmdInstanceModulesRemoveOptions.waitDone && cmdInstanceModulesRemoveOptions.moduleName != "cloud-manager" {
			return fmt.Errorf("--wait is only supported for --module cloud-manager")
		}

		if !isFound {
			fmt.Println("Module is already removed")
		} else {
			kyma.Spec.Modules = pie.FilterNot(kyma.Spec.Modules, func(m operatorv1beta2.Module) bool {
				return m.Name == cmdInstanceModulesRemoveOptions.moduleName
			})

			err = clnt.Update(rootCtx, kyma)
			if err != nil {
				return fmt.Errorf("failed to update SKR kyma: %w", err)
			}

			fmt.Println("Module is removed")
		}

		if cmdInstanceModulesRemoveOptions.waitDone {
			fmt.Printf("Waiting for cloud-manager finalizers to be removed with timeout %s\n", cmdInstanceModulesRemoveOptions.timeout.String())

			logger := logr.Discard()
			if verbose {
				logger = rootLogger.WithName("waitModuleRemove")
			}

			if pollErr := wait.PollUntilContextTimeout(rootCtx, 5*time.Second, cmdInstanceModulesRemoveOptions.timeout, false, func(ctx context.Context) (bool, error) {
				skrKyma := &operatorv1beta2.Kyma{}
				err := clnt.Get(ctx, types.NamespacedName{
					Namespace: "kyma-system",
					Name:      "default",
				}, skrKyma)
				if err != nil && !apierrors.IsNotFound(err) {
					return false, fmt.Errorf("failed to get SKR Kyma: %w", err)
				}
				if err == nil && controllerutil.ContainsFinalizer(skrKyma, api.CommonFinalizerDeletionHook) {
					logger.Info("SKR Kyma deletion-hook finalizer still present")
					return false, nil
				}

				cr := &cloudresourcesv1beta1.CloudResources{}
				err = clnt.Get(ctx, types.NamespacedName{
					Namespace: "kyma-system",
					Name:      "default",
				}, cr)
				if err != nil && !apierrors.IsNotFound(err) {
					return false, fmt.Errorf("failed to get SKR CloudResources: %w", err)
				}
				if err == nil && controllerutil.ContainsFinalizer(cr, api.CommonFinalizerDeletionHook) {
					logger.Info("SKR CloudResources deletion-hook finalizer still present")
					return false, nil
				}

				return true, nil
			}); pollErr != nil {
				fmt.Printf("Warning: poll exited early: %v\n", pollErr)
			}

			if err := forceRemoveCloudManagerFinalizers(clnt); err != nil {
				return err
			}

			fmt.Println("Cloud-manager finalizers removed")
		}

		return nil
	},
}

func init() {
	cmdInstanceModules.AddCommand(cmdInstanceModulesRemove)
	cmdInstanceModulesRemove.Flags().StringVarP(&cmdInstanceModulesRemoveOptions.runtimeID, "runtime-id", "r", "", "The runtime ID")
	cmdInstanceModulesRemove.Flags().StringVarP(&cmdInstanceModulesRemoveOptions.alias, "alias", "a", "", "The runtime alias")
	cmdInstanceModulesRemove.Flags().StringVarP(&cmdInstanceModulesRemoveOptions.moduleName, "module", "m", "", "The module name")
	cmdInstanceModulesRemove.Flags().BoolVarP(&cmdInstanceModulesRemoveOptions.waitDone, "wait", "w", false, "Wait until cloud-manager finalizers are removed (only supported for --module cloud-manager)")
	cmdInstanceModulesRemove.Flags().DurationVarP(&cmdInstanceModulesRemoveOptions.timeout, "timeout", "t", 5*time.Minute, "Timeout for waiting")
	_ = cmdInstanceModulesRemove.MarkFlagRequired("module")
	cmdInstanceModulesRemove.MarkFlagsMutuallyExclusive("runtime-id", "alias")
	cmdInstanceModulesRemove.MarkFlagsOneRequired("runtime-id", "alias")
}

func forceRemoveCloudManagerFinalizers(skrClient client.Client) error {
	skrKyma := &operatorv1beta2.Kyma{}
	err := skrClient.Get(rootCtx, types.NamespacedName{
		Namespace: "kyma-system",
		Name:      "default",
	}, skrKyma)
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("failed to get SKR Kyma for force finalizer removal: %w", err)
	}
	if err == nil && controllerutil.ContainsFinalizer(skrKyma, api.CommonFinalizerDeletionHook) {
		fmt.Printf("Force-removing deletion-hook finalizer from SKR Kyma\n")
		base := skrKyma.DeepCopyObject().(client.Object)
		controllerutil.RemoveFinalizer(skrKyma, api.CommonFinalizerDeletionHook)
		if err := skrClient.Patch(rootCtx, skrKyma, client.MergeFrom(base)); err != nil {
			return fmt.Errorf("failed to force-remove finalizer from SKR Kyma: %w", err)
		}
	}

	cr := &cloudresourcesv1beta1.CloudResources{}
	err = skrClient.Get(rootCtx, types.NamespacedName{
		Namespace: "kyma-system",
		Name:      "default",
	}, cr)
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("failed to get SKR CloudResources for force finalizer removal: %w", err)
	}
	if err == nil && controllerutil.ContainsFinalizer(cr, api.CommonFinalizerDeletionHook) {
		fmt.Printf("Force-removing deletion-hook finalizer from SKR CloudResources\n")
		base := cr.DeepCopyObject().(client.Object)
		controllerutil.RemoveFinalizer(cr, api.CommonFinalizerDeletionHook)
		if err := skrClient.Patch(rootCtx, cr, client.MergeFrom(base)); err != nil {
			return fmt.Errorf("failed to force-remove finalizer from SKR CloudResources: %w", err)
		}
	}

	return nil
}
