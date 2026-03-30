package sim

import (
	"context"
	"fmt"
	"time"

	"github.com/kyma-project/cloud-manager/api"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/external/operatorshared"
	"github.com/kyma-project/cloud-manager/pkg/external/operatorv1beta2"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/clock"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const timeoutAddModuleToReadyState = 30 * time.Second
const timeoutRemoveModuleToErrorState = 5 * time.Minute
const timeoutForceRemoveModule = 10 * time.Minute

func newSimKymaSkr(kcp client.Client, skr client.Client, runtimeID, kcpNamespace string, clck clock.Clock) *simKymaSkr {
	return &simKymaSkr{
		kcp:          kcp,
		skr:          skr,
		runtimeID:    runtimeID,
		kcpNamespace: kcpNamespace,
		clock:        clck,

		timeoutAddModuleToReadyState:    timeoutAddModuleToReadyState,
		timeoutRemoveModuleToErrorState: timeoutRemoveModuleToErrorState,
		timeoutForceRemoveModule:        timeoutForceRemoveModule,
	}
}

type simKymaSkr struct {
	kcp          client.Client
	skr          client.Client
	runtimeID    string
	kcpNamespace string
	clock        clock.Clock

	timeoutAddModuleToReadyState    time.Duration
	timeoutRemoveModuleToErrorState time.Duration
	timeoutForceRemoveModule        time.Duration
}

func (r *simKymaSkr) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	if request.String() != "kyma-system/default" {
		return reconcile.Result{}, nil
	}

	logger := composed.LoggerFromCtx(ctx)

	// load SKR Kyma ====================================================================

	skrKyma := &operatorv1beta2.Kyma{}
	err := r.skr.Get(ctx, request.NamespacedName, skrKyma)
	if client.IgnoreNotFound(err) != nil && util.IgnoreNoMatch(err) != nil {
		return reconcile.Result{}, fmt.Errorf("error getting Kyma: %w", err)
	}
	if err != nil {
		skrKyma = nil
	}

	// load CloudResources CM ==========================================================

	cm := &cloudresourcesv1beta1.CloudResources{}
	err = r.skr.Get(ctx, types.NamespacedName{
		Namespace: "kyma-system",
		Name:      "default",
	}, cm)
	if client.IgnoreNotFound(err) != nil && util.IgnoreNoMatch(err) != nil {
		return reconcile.Result{}, fmt.Errorf("error getting default CloudResources: %w", err)
	}
	if err != nil {
		cm = nil
	}

	// find runtimeId =================================================================

	var runtimeId string
	if skrKyma != nil {
		runtimeId = skrKyma.Labels[cloudcontrolv1beta1.LabelRuntimeId]
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
	}

	// load KCP Kyma =================================================================

	kcpKyma := &operatorv1beta2.Kyma{}
	err = r.kcp.Get(ctx, types.NamespacedName{
		Namespace: r.kcpNamespace,
		Name:      runtimeId,
	}, kcpKyma)
	if client.IgnoreNotFound(err) != nil && util.IgnoreNoMatch(err) != nil {
		return reconcile.Result{}, fmt.Errorf("error getting KCP Kyma %q: %w", runtimeId, err)
	}
	if err != nil {
		kcpKyma = nil
	}

	// handle loaded resources =====================================================

	if skrKyma == nil {
		return reconcile.Result{}, nil
	}
	if kcpKyma == nil {
		return reconcile.Result{}, nil
	}

	// determine the outcomes and change SKR Kyma status and KCP Kyma spec
	outcome := (&KymaSync{SKR: skrKyma, KCP: kcpKyma}).Sync()
	outcome.AutoProcessAllBut("cloud-manager")

	// delete ================================================================

	if skrKyma.DeletionTimestamp != nil {

		// disable all modules and save SKR Kyma
		// if cloud-manager module is enabled its status will go to Processing state
		if len(skrKyma.Spec.Modules) > 0 {
			logger.Info("Disabling all modules in SKR Kyma spec")
			skrKyma.Spec.Modules = nil
			outcome.SKR.SpecChanged = true
			if err := outcome.PatchObjects(ctx, r.skr, r.kcp); err != nil {
				return reconcile.Result{}, err
			}
			return reconcile.Result{RequeueAfter: util.Timing.T1000ms()}, nil
		}

		// CloudResources CR exists
		if cm != nil {

			// not deleted
			if cm.DeletionTimestamp == nil {
				// delete it
				logger.Info("Deleting default CloudResources")
				err = r.skr.Delete(ctx, cm)
				if client.IgnoreNotFound(err) != nil {
					return reconcile.Result{}, fmt.Errorf("error deleting default CloudResources: %w", err)
				}
				return reconcile.Result{RequeueAfter: util.Timing.T1000ms()}, nil
			}

			// wait deleted
			// CM should respond and eventually remove the finalizer

			elapsed := r.clock.Since(cm.CreationTimestamp.Time)

			if elapsed >= r.timeoutRemoveModuleToErrorState {
				ms := skrKyma.GetModuleStatusMap()["cloud-manager"]
				if ms.State != operatorshared.StateError {
					outcome.Processed("cloud-manager", operatorshared.StateError, "Timeout waiting for module to be deleted")

					util.ExpiringSwitch().
						Key("e2e.sim.kymaSkr.wait.CloudResources.remove.timeout", r.runtimeID).
						IfNotRecently(func() {
							logger.
								WithValues(
									"since", elapsed.String(),
									"timeout", r.timeoutAddModuleToReadyState.String(),
								).
								Info("Timeout waiting CloudResources CR to be deleted")
						})
				}
			} // if timeout

			if elapsed >= r.timeoutForceRemoveModule {
				util.ExpiringSwitch().
					Key("e2e.sim.kymaSkr.wait.CloudResources.force.remove.timeout", r.runtimeID).
					IfNotRecently(func() {
						logger.
							WithValues(
								"since", elapsed.String(),
								"timeout", r.timeoutAddModuleToReadyState.String(),
							).
							Info("Force removing CloudResources CR")
					})

				p := []byte(`[{"op": "remove", "path": "/metadata/finalizers"}]`)
				err = r.skr.Patch(ctx, cm, client.RawPatch(types.JSONPatchType, p))
				if err != nil {
					return reconcile.Result{}, fmt.Errorf("error removing CloudResources CR finalizers: %w", err)
				}

				outcome.Processed("cloud-manager", operatorshared.StateReady, "")
			}

			if err := outcome.PatchObjects(ctx, r.skr, r.kcp); err != nil {
				return reconcile.Result{}, err
			}

			return reconcile.Result{RequeueAfter: util.Timing.T1000ms()}, nil
		} // if CloudResources exists

		// CloudResources does not exist

		logger.Info("Removing SKR Kyma finalizer")
		_, err := composed.PatchObjRemoveFinalizer(ctx, api.CommonFinalizerDeletionHook, skrKyma, r.skr)
		if client.IgnoreNotFound(err) != nil && util.IgnoreNoMatch(err) != nil {
			return reconcile.Result{}, fmt.Errorf("error removing SKR Kyma finalizer: %w", err)
		}

		return reconcile.Result{}, nil
	}

	// create & update ========================================================

	// CloudManager is not enabled (not in spec) ===============================================================

	if !outcome.IsActive("cloud-manager") {

		if cm != nil {
			// CloudResources CR exists

			// if not deleted yet, delete it now
			if cm.DeletionTimestamp == nil {
				logger.Info("Deleting CloudResources")
				err = r.skr.Delete(ctx, cm)
				if util.IgnoreNoMatch(client.IgnoreNotFound(err)) != nil {
					return reconcile.Result{}, fmt.Errorf("error deleting CloudResources CR: %w", err)
				}
				return reconcile.Result{RequeueAfter: util.Timing.T1000ms()}, nil
			}

			// waiting to be deleted

			// check for timeout
			elapsed := r.clock.Since(cm.CreationTimestamp.Time)
			if elapsed >= r.timeoutRemoveModuleToErrorState {
				ms := skrKyma.GetModuleStatusMap()["cloud-manager"]
				if ms.State != operatorshared.StateError {
					outcome.Processed("cloud-manager", operatorshared.StateError, "Timeout waiting for module to be deleted")

					util.ExpiringSwitch().
						Key("e2e.sim.kymaSkr.wait.CloudResources.remove.timeout", r.runtimeID).
						IfNotRecently(func() {
							logger.
								WithValues(
									"since", elapsed.String(),
									"timeout", r.timeoutAddModuleToReadyState.String(),
								).
								Info("Timeout waiting CloudResources CR to be deleted")
						})

				}
			} // if timeout

			if err := outcome.PatchObjects(ctx, r.skr, r.kcp); err != nil {
				return reconcile.Result{}, err
			}

			return reconcile.Result{RequeueAfter: util.Timing.T1000ms()}, nil
		} // if exists

		// CloudResources CR do not exist

		outcome.Processed("cloud-manager", operatorshared.StateReady, "")

	} // if cloud-manager is disabled

	// CloudManager  is enabled (in spec) ===========================================================================

	if outcome.IsActive("cloud-manager") {

		// CM CR does not exist  => create CloudResources CM
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
			return reconcile.Result{RequeueAfter: util.Timing.T100ms()}, nil
		}

		// waiting to be to ready
		util.ExpiringSwitch().
			Key("e2e.sim.kymaSkr.wait.CloudResources.add", r.runtimeID).
			IfNotRecently(func() {
				logger.
					Info("Waiting CloudResources CR to become Ready")
			})

		// CM CR exists, check if ready

		if cm.Status.State != "Ready" {

			// CloudResources CR is not ready

			elapsed := r.clock.Since(cm.CreationTimestamp.Time)
			if elapsed >= r.timeoutAddModuleToReadyState {
				ms := skrKyma.GetModuleStatusMap()["cloud-manager"]
				if ms.State != operatorshared.StateError {
					outcome.Processed("cloud-manager", operatorshared.StateError, "Timeout waiting for module to be ready")

					util.ExpiringSwitch().
						Key("e2e.sim.kymaSkr.wait.CloudResources.add.timeout", r.runtimeID).
						IfNotRecently(func() {
							logger.
								WithValues(
									"since", elapsed.String(),
									"timeout", r.timeoutAddModuleToReadyState.String(),
								).
								Info("Timeout waiting CloudResources CR to become Ready")
						})
				}
			} // if timeout

			if err := outcome.PatchObjects(ctx, r.skr, r.kcp); err != nil {
				return reconcile.Result{}, err
			}

			return reconcile.Result{RequeueAfter: util.Timing.T1000ms()}, nil

		} // if not ready

		// CloudResources CR is ready
		// ensure SKR Kyma module status is Ready

		outcome.Processed("cloud-manager", operatorshared.StateReady, "Success")

	} // if cloud-manager is enabled

	// save SKR/KCP Kyma

	outcome.SetSkrKymaReadyStatus()

	if err := outcome.PatchObjects(ctx, r.skr, r.kcp); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *simKymaSkr) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named(fmt.Sprintf("kyma-skr-%s", r.runtimeID)).
		For(&operatorv1beta2.Kyma{}).
		Complete(r)
}
