package sim

import (
	"context"
	"fmt"
	"time"

	gardenertypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	gardenerconstants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	"github.com/kyma-project/cloud-manager/api"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	e2econfig "github.com/kyma-project/cloud-manager/e2e/config"
	e2elib "github.com/kyma-project/cloud-manager/e2e/lib"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"
	"github.com/kyma-project/cloud-manager/pkg/external/operatorshared"
	"github.com/kyma-project/cloud-manager/pkg/external/operatorv1beta2"
	"github.com/kyma-project/cloud-manager/pkg/util"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/clock"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func newSimRuntime(kcp client.Client, garden client.Client, cpl e2elib.CloudProfileLoader, config *e2econfig.ConfigType) *simRuntime {
	return &simRuntime{
		kcp:    kcp,
		garden: garden,
		config: config,
		cpl:    cpl,
		clock:  clock.RealClock{},
	}
}

var _ reconcile.Reconciler = &simRuntime{}

type simRuntime struct {
	kcp    client.Client
	garden client.Client
	config *e2econfig.ConfigType
	cpl    e2elib.CloudProfileLoader
	clock  clock.Clock
}

var GardenerConditionTypes = []gardenertypes.ConditionType{
	gardenertypes.ShootControlPlaneHealthy,
	gardenertypes.ShootAPIServerAvailable,
	gardenertypes.ShootEveryNodeReady,
	gardenertypes.ShootSystemComponentsHealthy,
}

func (r *simRuntime) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	result, err := r.reconcileRequest(ctx, request)
	//logger := composed.LoggerFromCtx(ctx)
	//if err != nil {
	//	logger.Error(err, "reconciliation failed with error")
	//} else if result.Requeue {
	//	logger.Info("reconciliation requeue")
	//} else if result.RequeueAfter > 0 {
	//	logger.Info(fmt.Sprintf("reconciliation delayed requeue after %s", result.RequeueAfter.String()))
	//} else {
	//	logger.Info("reconciliation succeeded")
	//}
	return result, err
}

func (r *simRuntime) reconcileRequest(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	logger := composed.LoggerFromCtx(ctx)

	rt := &infrastructuremanagerv1.Runtime{}
	err := r.kcp.Get(ctx, request.NamespacedName, rt)
	if client.IgnoreNotFound(err) != nil {
		return reconcile.Result{}, fmt.Errorf("error loading Runtime: %w", err)
	}
	if apierrors.IsNotFound(err) {
		return reconcile.Result{}, nil
	}

	if _, ok := rt.Labels[e2elib.DoNotReconcile]; ok {
		return reconcile.Result{}, nil
	}

	logger = logger.WithValues("shootName", rt.Spec.Shoot.Name)

	shoot := &gardenertypes.Shoot{}
	err = r.garden.Get(ctx, types.NamespacedName{
		Namespace: r.config.GardenNamespace,
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
			return reconcile.Result{RequeueAfter: util.Timing.T10000ms()}, nil
		}
		if kyma != nil && kyma.DeletionTimestamp != nil {
			// waiting for kyma to get deleted
			util.ExpiringSwitch().
				Key("sim.runtime.wait.kcp.kyma.deleted", request.NamespacedName.String()).
				IfNotRecently(func() {
					logger.Info("Waiting KCP Kyma to be deleted")
				})
			diff := r.clock.Now().Sub(kyma.DeletionTimestamp.Time)
			if diff < time.Minute {
				return reconcile.Result{RequeueAfter: util.Timing.T10000ms()}, nil
			}
			logger.Info("Timeout on waiting KCP Kyma to be deleted after 1 minute")
		}

		if shoot != nil && shoot.DeletionTimestamp.IsZero() {
			logger.Info("Deleting Shoot")
			_, err = composed.PatchObjMergeAnnotation(ctx, gardenerconstants.ConfirmationDeletion, "True", shoot, r.garden)
			if err != nil {
				return reconcile.Result{}, fmt.Errorf("error addind confirmationDeletion annotation on the shoot: %w", err)
			}
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
			util.ExpiringSwitch().
				Key("sim.runtime.wait.shoot.deleted", request.NamespacedName.String()).
				IfNotRecently(func() {
					logger.Info("Waiting shoot to be deleted...")
				})
			return reconcile.Result{RequeueAfter: util.Timing.T10000ms()}, nil
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

	finalizerAdded, err := composed.PatchObjAddFinalizer(ctx, infrastructuremanagerv1.Finalizer, rt, r.kcp)
	if err != nil {
		logger.Error(err, "Error adding finalizer")
		return reconcile.Result{}, fmt.Errorf("error adding Runtime finalizer: %w", err)
	}
	if finalizerAdded {
		logger.Info("Added finalizer to Runtime")
		return reconcile.Result{RequeueAfter: util.Timing.T1000ms()}, nil
	}

	if shoot == nil {
		//logger.Info("Shoot not found")
		cpr, err := r.cpl.Load(ctx)
		if err != nil {
			logger.Error(err, "Error loading CloudProfiles")
			return reconcile.Result{}, err
		}
		shootBuilder := NewShootBuilder(cpr, r.config).
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
		shootList := &gardenertypes.ShootList{}
		err = r.garden.List(ctx, shootList, client.InNamespace(r.config.GardenNamespace))
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("error listing shoots in Garden: %w", err)
		}

		logger.
			WithValues(
				"shootName", shoot.Name,
				"shootNamespace", shoot.Namespace,
			).
			Info("Creating Shoot")
		err = r.garden.Create(ctx, shoot)
		if apierrors.IsAlreadyExists(err) {
			return reconcile.Result{RequeueAfter: util.Timing.T1000ms()}, nil
		}
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("error creating Shoot: %w", err)
		}

		return reconcile.Result{RequeueAfter: util.Timing.T10000ms()}, nil
	}

	//logger.Info("Shoot is loaded")

	if len(shoot.Status.LastErrors) > 0 {
		statusChanged := false
		if rt.Status.ProvisioningCompleted {
			rt.Status.ProvisioningCompleted = false
			statusChanged = true
		}
		if len(rt.Status.Conditions) != 1 {
			rt.Status.Conditions = []metav1.Condition{
				{
					Type:               cloudcontrolv1beta1.ConditionTypeError,
					Status:             metav1.ConditionTrue,
					ObservedGeneration: rt.Generation,
					LastTransitionTime: metav1.Time{Time: r.clock.Now()},
					Reason:             "ShootError",
					Message:            shoot.Status.LastErrors[0].Description,
				},
			}
			statusChanged = true
		}
		if rt.Status.State != infrastructuremanagerv1.RuntimeStateFailed {
			rt.Status.State = infrastructuremanagerv1.RuntimeStateFailed
			statusChanged = true
		}

		if statusChanged {
			logger.Info("Updating Runtime Status with shoot error state")
			err = composed.PatchObjStatus(ctx, rt, r.kcp)
			if err != nil {
				return reconcile.Result{}, fmt.Errorf("error patching Runtime status with shoot error state: %w", err)
			}
		}

		return reconcile.Result{RequeueAfter: util.Timing.T10000ms()}, nil
	} else {
		statusChanged := false
		if len(rt.Status.Conditions) != 0 {
			rt.Status.Conditions = []metav1.Condition{}
			statusChanged = true
		}
		if rt.Status.State == infrastructuremanagerv1.RuntimeStateFailed || rt.Status.State == "" {
			rt.Status.State = infrastructuremanagerv1.RuntimeStatePending
			statusChanged = true
		}
		if statusChanged {
			logger.Info("Updating Runtime Status with pending state")
			err = composed.PatchObjStatus(ctx, rt, r.kcp)
			if err != nil {
				return reconcile.Result{}, fmt.Errorf("error patching Runtime status with pending state: %w", err)
			}
			return reconcile.Result{RequeueAfter: util.Timing.T1000ms()}, nil
		}
	}

	if !IsShootReady(shoot) {
		util.ExpiringSwitch().
			Key("sim.runtime.wait.shoot.ready", request.NamespacedName.String()).
			IfNotRecently(func() {
				logger.Info("Waiting shoot to become ready...")
			})
		return reconcile.Result{RequeueAfter: util.Timing.T1000ms()}, nil
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
		util.ExpiringSwitch().
			Key("sim.runtime.wait.gardenerCluster.ready", request.NamespacedName.String()).
			IfNotRecently(func() {
				logger.Info("Waiting for GardenCluster to become ready")
			})
		return reconcile.Result{RequeueAfter: util.Timing.T10000ms()}, nil
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
		util.ExpiringSwitch().
			Key("sim.runtime.wait.kcp.kyma.ready", request.NamespacedName.String()).
			IfNotRecently(func() {
				logger.Info("Waiting for KCP Kyma to become ready")
			})
		return reconcile.Result{RequeueAfter: util.Timing.T10000ms()}, nil
	}

	_, err = composed.PatchObjAddFinalizer(ctx, infrastructuremanagerv1.Finalizer, rt, r.kcp)
	if err != nil {
		logger.Error(err, "Failed to add finalizer to Runtime")
		return reconcile.Result{}, fmt.Errorf("failed to add finalizer to Runtime: %w", err)
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
		logger.Info("Updating Runtime Status with ready state")
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
