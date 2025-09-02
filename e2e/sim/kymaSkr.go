package sim

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/external/operatorv1beta2"
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

	kcpKyma := &operatorv1beta2.Kyma{}
	err = r.kcp.Get(ctx, request.NamespacedName, kcpKyma)

	// todo

}

func (r *simKymaSkr) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1beta2.Kyma{}).
		Complete(r)
}
