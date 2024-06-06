package testinfra

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
	reconcile2 "github.com/kyma-project/cloud-manager/pkg/skr/runtime/reconcile"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type pvcControllerFactory struct{}

func (f *pvcControllerFactory) New(args reconcile2.ReconcilerArguments) reconcile.Reconciler {
	return &pvcController{
		skrCluster: args.SkrCluster,
	}
}

type pvcController struct {
	skrCluster cluster.Cluster
}

func (c *pvcController) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	pvc := &corev1.PersistentVolumeClaim{}
	err := c.skrCluster.GetClient().Get(ctx, request.NamespacedName, pvc)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	isLabeled := false
	if pvc.Labels != nil {
		_, isLabeled = pvc.Labels[cloudresourcesv1beta1.LabelCloudManaged]
	}
	if !isLabeled {
		return ctrl.Result{}, nil
	}

	if pvc.DeletionTimestamp.IsZero() && pvc.Status.Phase == corev1.ClaimPending {
		pvc.Status.Phase = corev1.ClaimBound
		err = c.skrCluster.GetClient().Status().Update(ctx, pvc)
		return ctrl.Result{}, err
	}

	if !pvc.DeletionTimestamp.IsZero() && pvc.Status.Phase == corev1.ClaimBound {
		controllerutil.RemoveFinalizer(pvc, "kubernetes.io/pvc-protection")
		err = c.skrCluster.GetClient().Update(ctx, pvc)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func SetupPVCController(reg skrruntime.SkrRegistry) error {
	return reg.Register().
		WithFactory(&pvcControllerFactory{}).
		For(&corev1.PersistentVolumeClaim{}).
		Complete()
}
