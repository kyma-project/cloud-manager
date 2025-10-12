package sim

import (
	"context"
	"fmt"
	"time"

	gardenertypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	gardenerhelper "github.com/gardener/gardener/pkg/apis/core/v1beta1/helper"
	"github.com/kyma-project/cloud-manager/api"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	e2econfig "github.com/kyma-project/cloud-manager/e2e/config"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"
	"github.com/kyma-project/cloud-manager/pkg/external/operatorshared"
	"github.com/kyma-project/cloud-manager/pkg/external/operatorv1beta2"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/clock"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func newSimRuntime(kcp client.Client, garden client.Client, cpl CloudProfileLoader) *simRuntime {
	return &simRuntime{
		kcp:    kcp,
		garden: garden,
		cpl:    cpl,
		clock:  clock.RealClock{},
	}
}

var _ reconcile.Reconciler = &simRuntime{}

type simRuntime struct {
	kcp    client.Client
	garden client.Client
	cpl    CloudProfileLoader
	clock  clock.Clock
}

var GardenerConditionTypes = []gardenertypes.ConditionType{
	gardenertypes.ShootControlPlaneHealthy,
	gardenertypes.ShootAPIServerAvailable,
	gardenertypes.ShootEveryNodeReady,
	gardenertypes.ShootSystemComponentsHealthy,
}

func (r *simRuntime) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	logger := composed.LoggerFromCtx(ctx)

	rt := &infrastructuremanagerv1.Runtime{}
	err := r.kcp.Get(ctx, request.NamespacedName, rt)
	if client.IgnoreNotFound(err) != nil {
		return reconcile.Result{}, fmt.Errorf("error loading Runtime: %w", err)
	}
	if apierrors.IsNotFound(err) {
		return reconcile.Result{}, nil
	}

	if _, ok := rt.Labels[DoNotReconcile]; ok {
		return reconcile.Result{}, nil
	}

	shoot := &gardenertypes.Shoot{}
	err = r.garden.Get(ctx, types.NamespacedName{
		Namespace: e2econfig.Config.GardenNamespace,
		Name:      rt.Spec.Shoot.Name,
	}, shoot)
	if client.IgnoreNotFound(err) != nil {
		return reconcile.Result{}, fmt.Errorf("error loading Shoot: %w", err)
	}
	if apierrors.IsNotFound(err) {
		shoot = nil
	}

	gc := &infrastructuremanagerv1.GardenerCluster{}
	err = r.kcp.Get(ctx, request.NamespacedName, gc)
	if client.IgnoreNotFound(err) != nil {
		return reconcile.Result{}, fmt.Errorf("error loading GardenerCluster: %w", err)
	}
	if apierrors.IsNotFound(err) {
		gc = nil
	}

	kyma := &operatorv1beta2.Kyma{}
	err = r.kcp.Get(ctx, request.NamespacedName, kyma)
	if client.IgnoreNotFound(err) != nil {
		return reconcile.Result{}, fmt.Errorf("error loading KCP Kyma: %w", err)
	}
	if apierrors.IsNotFound(err) {
		kyma = nil
	}

	// delete ==========================================

	if rt.DeletionTimestamp != nil {
		if kyma != nil && kyma.DeletionTimestamp.IsZero() {
			logger.Info("Deleting KCP Kyma")
			err = r.kcp.Delete(ctx, kyma)
			if err != nil {
				return reconcile.Result{}, fmt.Errorf("error deleting KCP Kyma: %w", err)
			}
			return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
		}
		if kyma != nil && kyma.DeletionTimestamp != nil {
			// waiting for kyma to get deleted
			diff := r.clock.Now().Sub(kyma.DeletionTimestamp.Time)
			if diff < time.Minute {
				logger.Info("Waiting KCP Kyma gets deleted")
				return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
			}
			logger.Info("Giving up on waiting KCP Kyma gets deleted after 1 minute")
		}

		if shoot != nil && shoot.DeletionTimestamp.IsZero() {
			logger.Info("Deleting Shoot")
			err = r.garden.Delete(ctx, shoot)
			if err != nil {
				return reconcile.Result{}, fmt.Errorf("error deleting Shoot: %w", err)
			}
		}

		if gc != nil && gc.DeletionTimestamp.IsZero() {
			logger.Info("Deleting GardenCluster")
			err = r.kcp.Delete(ctx, gc)
			if err != nil {
				return reconcile.Result{}, fmt.Errorf("error deleting GardenerCluster: %w", err)
			}
		}

		if shoot != nil || gc != nil {
			return reconcile.Result{RequeueAfter: 10 * time.Second}, nil
		}

		logger.Info("Removing Runtime finalizer")
		_, err = composed.PatchObjRemoveFinalizer(ctx, infrastructuremanagerv1.Finalizer, rt, r.kcp)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("error removing Runtime finalizer: %w", err)
		}
		return reconcile.Result{}, nil
	}

	// TODO once implemented, load the VpcNetwork referenced from the Runtime and wait until Ready if `type: kyma`

	// create ====================================================

	if shoot == nil {
		cpr, err := r.cpl.Load(ctx)
		if err != nil {
			logger.Error(err, "Error loading CloudProfiles")
			return reconcile.Result{}, err
		}
		shootBuilder := NewShootBuilder(cpr).
			WithRuntime(rt)
		if errShoot := shootBuilder.Validate(); errShoot != nil {
			rt.Status.State = infrastructuremanagerv1.ErrorState
			meta.SetStatusCondition(&rt.Status.Conditions, metav1.Condition{
				Type:               string(infrastructuremanagerv1.ConditionTypeRuntimeProvisioned),
				Status:             metav1.ConditionFalse,
				ObservedGeneration: rt.Generation,
				Reason:             string(infrastructuremanagerv1.ConditionReasonCreationError),
				Message:            fmt.Sprintf("Shoot validation error: %s", errShoot.Error()),
			})
			err = composed.PatchObjStatus(ctx, rt, r.kcp)
			if err != nil {
				return reconcile.Result{}, fmt.Errorf("error patching Runtime status with shoot validation error (%s): %s", errShoot, err)
			}
			return reconcile.Result{}, nil
		}
		shoot = shootBuilder.Build()
		logger.Info("Creating Shoot")
		err = r.garden.Create(ctx, shoot)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("error creating Shoot: %w", err)
		}

		return reconcile.Result{RequeueAfter: 30 * time.Second}, nil
	}

	if len(shoot.Status.Conditions) == 0 {
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	}
	for _, ct := range GardenerConditionTypes {
		cond := gardenerhelper.GetCondition(shoot.Status.Conditions, ct)
		if cond == nil {
			return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
		}
		if cond.Status != gardenertypes.ConditionTrue {
			return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
		}
	}

	// shoot is ready

	if gc == nil {

		gc = &infrastructuremanagerv1.GardenerCluster{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: rt.Namespace,
				Name:      rt.Name,
				Labels: map[string]string{
					cloudcontrolv1beta1.LabelScopeGlobalAccountId: rt.Labels[cloudcontrolv1beta1.LabelScopeGlobalAccountId],
					cloudcontrolv1beta1.LabelScopeSubaccountId:    rt.Labels[cloudcontrolv1beta1.LabelScopeSubaccountId],
					cloudcontrolv1beta1.LabelScopeShootName:       shoot.Name,
					cloudcontrolv1beta1.LabelScopeRegion:          rt.Labels[cloudcontrolv1beta1.LabelScopeRegion],
					cloudcontrolv1beta1.LabelScopeBrokerPlanName:  rt.Labels[cloudcontrolv1beta1.LabelScopeBrokerPlanName],
					cloudcontrolv1beta1.LabelScopeProvider:        rt.Labels[cloudcontrolv1beta1.LabelScopeProvider],
					cloudcontrolv1beta1.LabelRuntimeId:            rt.Name,
				},
			},
			Spec: infrastructuremanagerv1.GardenerClusterSpec{
				Shoot: infrastructuremanagerv1.Shoot{
					Name: shoot.Name,
				},
				Kubeconfig: infrastructuremanagerv1.Kubeconfig{
					Secret: infrastructuremanagerv1.Secret{
						Key:       "config",
						Name:      fmt.Sprintf("kubeconfig-%s", rt.Name),
						Namespace: rt.Namespace,
					},
				},
			},
		}
		logger.Info("Creating GardenCluster")
		err = r.kcp.Create(ctx, gc)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("error creating gardencluster: %w", err)
		}
	}

	if gc.Status.State != infrastructuremanagerv1.ReadyState {
		logger.Info("Waiting for GardenCluster to become ready")
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	}

	if kyma == nil {
		kyma = &operatorv1beta2.Kyma{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: rt.Namespace,
				Name:      rt.Name,
				Labels:    rt.Labels,
				Finalizers: []string{
					api.CommonFinalizerDeletionHook,
				},
			},
			Spec: operatorv1beta2.KymaSpec{
				Channel: operatorv1beta2.DefaultChannel,
			},
		}
		kyma.Labels[cloudcontrolv1beta1.LabelRuntimeId] = rt.Name
		logger.Info("Creating KCP Kyma")
		err = r.kcp.Create(ctx, kyma)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("error creating Kyma: %w", err)
		}
	}

	if kyma.Status.State != operatorshared.StateReady {
		logger.Info("Waiting for Kyma to become ready")
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	}

	statusChanged := false
	if !rt.Status.ProvisioningCompleted {
		rt.Status.ProvisioningCompleted = true
		statusChanged = true
	}
	if len(rt.Status.Conditions) != 0 {
		rt.Status.Conditions = []metav1.Condition{}
		statusChanged = true
	}
	if rt.Status.State != infrastructuremanagerv1.ReadyState {
		rt.Status.State = infrastructuremanagerv1.ReadyState
		statusChanged = true
	}

	if statusChanged {
		logger.Info("Updating Runtime Status")
		err = composed.PatchObjStatus(ctx, rt, r.kcp)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("error patching Runtime status with ready state: %w", err)
		}
	}

	return reconcile.Result{}, nil
}

func (r *simRuntime) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrastructuremanagerv1.Runtime{}).
		Complete(r)
}
