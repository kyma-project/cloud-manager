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

type pvControllerFactory struct{}

func (f *pvControllerFactory) New(args reconcile2.ReconcilerArguments) reconcile.Reconciler {
	return &pvController{
		skrCluster: args.SkrCluster,
	}
}

type pvController struct {
	skrCluster cluster.Cluster
}

func (c *pvController) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	pv := &corev1.PersistentVolume{}
	err := c.skrCluster.GetClient().Get(ctx, request.NamespacedName, pv)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if pv.Spec.NFS == nil {
		return ctrl.Result{}, nil
	}
	isLabeled := false
	if pv.Labels != nil {
		_, isLabeled = pv.Labels[cloudresourcesv1beta1.LabelCloudManaged]
	}
	if !isLabeled {
		return ctrl.Result{}, nil
	}

	if pv.DeletionTimestamp.IsZero() && pv.Status.Phase == corev1.VolumePending {
		pv.Status.Phase = corev1.VolumeAvailable
		err = c.skrCluster.GetClient().Status().Update(ctx, pv)
		return ctrl.Result{}, err
	}

	if !pv.DeletionTimestamp.IsZero() && pv.Status.Phase == corev1.VolumeAvailable {
		controllerutil.RemoveFinalizer(pv, "kubernetes.io/pv-protection")
		err = c.skrCluster.GetClient().Update(ctx, pv)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func SetupPvController(reg skrruntime.SkrRegistry) error {
	return reg.Register().
		WithFactory(&pvControllerFactory{}).
		For(&corev1.PersistentVolume{}).
		Complete()
}
