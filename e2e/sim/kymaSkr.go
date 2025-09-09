package sim

import (
	"context"
	"fmt"

	"github.com/elliotchance/pie/v2"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	e2econfig "github.com/kyma-project/cloud-manager/e2e/config"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/external/operatorshared"
	"github.com/kyma-project/cloud-manager/pkg/external/operatorv1beta2"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func NewSimKymaSkr(mngr manager.Manager, kcp client.Client, skr client.Client) error {
	r := &simKymaSkr{
		kcp: kcp,
		skr: skr,
	}
	return r.SetupWithManager(mngr)
}

type simKymaSkr struct {
	kcp client.Client
	skr client.Client
}

func (r *simKymaSkr) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	if request.String() != "kyma-system/kyma" {
		return reconcile.Result{}, nil
	}

	skrKyma := &operatorv1beta2.Kyma{}
	err := r.skr.Get(ctx, request.NamespacedName, skrKyma)
	if client.IgnoreNotFound(err) != nil {
		return reconcile.Result{}, fmt.Errorf("error getting Kyma: %w", err)
	}
	if err != nil {
		return reconcile.Result{}, nil
	}

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

	// TODO: continue here - make two kymas wrapper that knows to sync modules in spec and status so it can be tested
	kcpKymaChanged := false
	modulesToAdd := r.moduleDifference(skrKyma.Spec.Modules, kcpKyma.Spec.Modules)
	if len(modulesToAdd) > 0 {
		kcpKymaChanged = true
		for _, m := range modulesToAdd {
			kcpKyma.Spec.Modules = append(kcpKyma.Spec.Modules, m)
			kcpKyma.Status.Modules = append(kcpKyma.Status.Modules, operatorv1beta2.ModuleStatus{
				Name:   m.Name,
				State:  operatorshared.StateReady,
			})
		}
	}

	modulesToRemove := r.moduleDifference(kcpKyma.Spec.Modules, skrKyma.Spec.Modules)
	if len(modulesToRemove) > 0 {
		kcpKymaChanged = true
		removeIndex := make(map[string]struct{}, len(modulesToRemove))
		for _, m := range modulesToRemove {
			removeIndex[m.Name] = struct{}{}
		}
		kcpKyma.Spec.Modules = pie.FilterNot(kcpKyma.Spec.Modules, func(m operatorv1beta2.Module) bool {
			_, ok := removeIndex[m.Name]
			return ok
		})
		kcpKyma.Status.Modules = pie.FilterNot(kcpKyma.Status.Modules, func(m operatorv1beta2.ModuleStatus) bool {
			_, ok := removeIndex[m.Name]
			return ok
		})
	}



	return reconcile.Result{}, err
}

func (r *simKymaSkr) syncKymaStatusModules(kyma *operatorv1beta2.Kyma) (bool) {
	changed := false
	index := make(map[string]struct{}, len(kyma.Spec.Modules))
	for _, m := range kyma.Spec.Modules {
		index[m.Name] = struct{}{}
	}
	kyma.Status.Modules
	for
}

func (r *simKymaSkr) moduleDifference(a []operatorv1beta2.Module, b []operatorv1beta2.Module) []operatorv1beta2.Module {
	var result []operatorv1beta2.Module
	for _, modA := range a {
		found := false
		for _, modB := range b {
			if modA.Name == modB.Name {
				// modA exists in b
				found = true
				break
			}
		}
		if !found {
			result = append(result, modA)
		}
	}
	return result
}

func (r *simKymaSkr) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("kyma-skr").
		For(&operatorv1beta2.Kyma{}).
		Complete(r)
}
