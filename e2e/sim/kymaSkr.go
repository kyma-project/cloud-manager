package sim

import (
	"context"
	"fmt"
	"time"

	"github.com/kyma-project/cloud-manager/api"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	e2econfig "github.com/kyma-project/cloud-manager/e2e/config"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/external/operatorshared"
	"github.com/kyma-project/cloud-manager/pkg/external/operatorv1beta2"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func NewSimKymaSkr(kcp client.Client, skr client.Client) *simKymaSkr {
	return &simKymaSkr{
		kcp: kcp,
		skr: skr,
	}
}

type simKymaSkr struct {
	kcp client.Client
	skr client.Client
}

func (r *simKymaSkr) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	if request.String() != "kyma-system/kyma" {
		return reconcile.Result{}, nil
	}

	logger := composed.LoggerFromCtx(ctx)

	skrKyma := &operatorv1beta2.Kyma{}
	err := r.skr.Get(ctx, request.NamespacedName, skrKyma)
	if client.IgnoreNotFound(err) != nil {
		return reconcile.Result{}, fmt.Errorf("error getting Kyma: %w", err)
	}
	if err != nil {
		return reconcile.Result{}, nil
	}

	// delete ================================================================

	if skrKyma.DeletionTimestamp != nil {
		cm := &cloudresourcesv1beta1.CloudResources{}
		err = r.skr.Get(ctx, types.NamespacedName{
			Namespace: "kyma-system",
			Name:      "default",
		}, cm)
		if client.IgnoreNotFound(err) != nil {
			return reconcile.Result{}, fmt.Errorf("error getting default CloudResources: %w", err)
		}
		if err != nil {
			cm = nil
		}

		if cm != nil && cm.DeletionTimestamp == nil {
			logger.Info("Deleting default CloudResources")
			err = r.skr.Delete(ctx, cm)
			if client.IgnoreNotFound(err) != nil {
				return reconcile.Result{}, fmt.Errorf("error deleting default CloudResources: %w", err)
			}
			return reconcile.Result{RequeueAfter: time.Second}, nil
		}

		if cm != nil {
			logger.Info("Waiting CloudResources are deleted...")
			return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
		}

		logger.Info("Removing SKR Kyma finalizer")
		_, err := composed.PatchObjRemoveFinalizer(ctx, api.CommonFinalizerDeletionHook, skrKyma, r.skr)
		if client.IgnoreNotFound(err) != nil {
			return reconcile.Result{}, fmt.Errorf("error removing SKR Kyma finalizer: %w", err)
		}

		return reconcile.Result{}, nil
	}

	// create & update ========================================================

	runtimeId := skrKyma.Labels[cloudcontrolv1beta1.LabelRuntimeId]
	if runtimeId == "" {
		statusChanged := false
		if skrKyma.Status.State != operatorshared.StateError {
			skrKyma.Status.State = operatorshared.StateError
			statusChanged = true
		}
		errCond := meta.FindStatusCondition(skrKyma.Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
		if errCond == nil || len(skrKyma.Status.Conditions) != 1 {
			meta.SetStatusCondition(&skrKyma.Status.Conditions, metav1.Condition{
				Type:               cloudcontrolv1beta1.ConditionTypeError,
				Status:             metav1.ConditionTrue,
				ObservedGeneration: skrKyma.Generation,
				Reason:             "NoRuntimeIdLabel",
				Message:            "Missing runtime ID label",
			})
			statusChanged = true
		}
		if statusChanged {
			logger.Info("Missing runtimeID label in SKR KCP")
			err = composed.PatchObjStatus(ctx, skrKyma, r.skr)
			if err != nil {
				return reconcile.Result{}, fmt.Errorf("error updating Kyma status for missing runtime id label: %w", err)
			}
		}
		return reconcile.Result{}, nil
	}

	kcpKyma := &operatorv1beta2.Kyma{}
	err = r.kcp.Get(ctx, types.NamespacedName{
		Namespace: e2econfig.Config.KcpNamespace,
		Name:      runtimeId,
	}, kcpKyma)
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("error getting KCP Kyma %q: %w", runtimeId, err)
	}

	outcome := (KymaSync{SKR: skrKyma, KCP: kcpKyma}).Sync()
	if outcome.SKR.StatusChanged {
		logger.Info("Patching SKR Kyma status.modules")
		if err := composed.PatchObjStatus(ctx, skrKyma, r.skr); err != nil {
			return reconcile.Result{}, fmt.Errorf("error patching SKR Kyma status: %w", err)
		}
	}
	if outcome.KCP.SpecChanged {
		logger.Info("Updating KCP Kyma spec.modules")
		if err := r.kcp.Update(ctx, kcpKyma); err != nil {
			return reconcile.Result{}, fmt.Errorf("error updating KCP Kyma spec: %w", err)
		}
	}
	if outcome.KCP.StatusChanged {
		logger.Info("Patching KCP Kyma status.modules")
		if err := composed.PatchObjStatus(ctx, kcpKyma, r.kcp); err != nil {
			return reconcile.Result{}, fmt.Errorf("error updating KCP Kyma status: %w", err)
		}
	}

	cm := &cloudresourcesv1beta1.CloudResources{}
	err = r.skr.Get(ctx, types.NamespacedName{
		Namespace: "kyma-system",
		Name:      "default",
	}, cm)
	if client.IgnoreNotFound(err) != nil {
		return reconcile.Result{}, fmt.Errorf("error getting default CloudResources: %w", err)
	}
	if err != nil {
		cm = nil
	}

	if cm == nil {
		logger.Info("Creating default CloudResources")
		cm = &cloudresourcesv1beta1.CloudResources{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "kyma-system",
				Name:      "default",
			},
		}
		err = r.skr.Create(ctx, cm)
		if client.IgnoreAlreadyExists(err) != nil {
			return reconcile.Result{}, fmt.Errorf("error creating CloudResources: %w", err)
		}
	}

	return reconcile.Result{}, nil
}

func (r *simKymaSkr) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("kyma-skr").
		For(&operatorv1beta2.Kyma{}).
		Complete(r)
}
