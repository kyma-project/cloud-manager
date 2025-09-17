package sim

import (
	"context"
	"fmt"
	"time"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/clock"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	expiresAtAnnotation               = "operator.kyma-project.io/expires-at"
	forceKubeconfigRotationAnnotation = "operator.kyma-project.io/force-kubeconfig-rotation"
)

func newSimGardenerCluster(kcp client.Client, kubeconfigProvider KubeconfigProvider) *simGardenerCluster {
	return &simGardenerCluster{
		kcp:                kcp,
		kubeconfigProvider: kubeconfigProvider,
		clock:              clock.RealClock{},
	}
}

var _ reconcile.Reconciler = &simGardenerCluster{}

type simGardenerCluster struct {
	kcp                client.Client
	kubeconfigProvider KubeconfigProvider
	clock              clock.Clock
}

func (r *simGardenerCluster) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	gc := &infrastructuremanagerv1.GardenerCluster{}
	err := r.kcp.Get(ctx, request.NamespacedName, gc)
	if apierrors.IsNotFound(err) {
		return reconcile.Result{}, nil
	}
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("error loading GardenerCluster: %w", err)
	}

	if _, ok := gc.Labels[DoNotReconcile]; ok {
		return reconcile.Result{}, nil
	}

	shootName := gc.Labels[cloudcontrolv1beta1.LabelScopeShootName]

	if shootName == "" {
		gc.Status.State = infrastructuremanagerv1.ErrorState
		meta.SetStatusCondition(&gc.Status.Conditions, metav1.Condition{
			Type:               string(infrastructuremanagerv1.ConditionTypeRuntimeKubeconfigReady),
			Status:             metav1.ConditionFalse,
			ObservedGeneration: gc.Generation,
			Reason:             string(infrastructuremanagerv1.ConditionReasonConfigurationErr),
			Message:            fmt.Sprintf("Runtime missing label %s", cloudcontrolv1beta1.LabelScopeShootName),
		})
		err = composed.PatchObjStatus(ctx, gc, r.kcp)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("error patching GardenerCluster status with missing cluster name label condition: %w", err)
		}
		return reconcile.Result{}, nil
	}

	syncNeeded, requeueAfter := r.isSyncNeeded(gc)
	if !syncNeeded {
		return reconcile.Result{RequeueAfter: requeueAfter}, nil
	}

	kubeConfigBytes, err := r.kubeconfigProvider.CreateNewKubeconfig(ctx, shootName)
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("error creating kubeconfig: %w", err)
	}

	kubeSecret := &corev1.Secret{}
	err = r.kcp.Get(ctx, types.NamespacedName{
		Namespace: gc.Spec.Kubeconfig.Secret.Namespace,
		Name:      gc.Spec.Kubeconfig.Secret.Name,
	}, kubeSecret)
	if client.IgnoreNotFound(err) != nil {
		return reconcile.Result{}, fmt.Errorf("error getting kubeconfig secret: %w", err)
	}
	if err != nil {
		// not found, create it
		kubeSecret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: gc.Spec.Kubeconfig.Secret.Namespace,
				Name:      gc.Spec.Kubeconfig.Secret.Name,
			},
			StringData: map[string]string{
				"config": string(kubeConfigBytes),
			},
		}
		err = r.kcp.Create(ctx, kubeSecret)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("error creating kubeconfig secret: %w", err)
		}
	} else {
		// found, update it
		kubeSecret.StringData = map[string]string{
			"config": string(kubeConfigBytes),
		}
		err = r.kcp.Update(ctx, kubeSecret)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("error updating kubeconfig secret: %w", err)
		}
	}

	if gc.Annotations == nil {
		gc.Annotations = map[string]string{}
	}
	gc.Annotations[expiresAtAnnotation] = r.clock.Now().Add(r.kubeconfigProvider.ExpiresIn()).Format(time.RFC3339)
	delete(gc.Annotations, forceKubeconfigRotationAnnotation)

	err = r.kcp.Update(ctx, gc)
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("error updating GardenerCluster annotations: %w", err)
	}

	statusChanged := false
	if gc.Status.State != infrastructuremanagerv1.ReadyState {
		gc.Status.State = infrastructuremanagerv1.ReadyState
		statusChanged = true
	}
	if len(gc.Status.Conditions) > 0 {
		gc.Status.Conditions = []metav1.Condition{}
		statusChanged = true
	}
	if statusChanged {
		err = composed.PatchObjStatus(ctx, gc, r.kcp)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("error patching GardenerCluster status with status condition: %w", err)
		}
	}

	return reconcile.Result{RequeueAfter: 10 * time.Minute}, nil
}

func (r *simGardenerCluster) isSyncNeeded(gc *infrastructuremanagerv1.GardenerCluster) (bool, time.Duration) {
	if gc.Status.State != infrastructuremanagerv1.ReadyState {
		return true, time.Minute
	}
	if gc.Annotations == nil {
		return true, time.Minute
	}
	if _, ok := gc.Annotations[forceKubeconfigRotationAnnotation]; ok {
		return true, time.Minute
	}
	expiresAt := r.clock.Now()
	val, ok := gc.Annotations[expiresAtAnnotation]
	if ok {
		ea, err := time.Parse(time.RFC3339, val)
		if err == nil {
			expiresAt = ea
		}
	}

	expiresIn := time.Until(expiresAt)
	if expiresIn < time.Hour {
		return true, time.Minute
	}

	return false, time.Duration(0.5 * float64(expiresIn))
}

func (r *simGardenerCluster) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrastructuremanagerv1.GardenerCluster{}).
		Complete(r)
}
