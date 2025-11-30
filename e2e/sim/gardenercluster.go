package sim

import (
	"context"
	"fmt"
	"time"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	e2elib "github.com/kyma-project/cloud-manager/e2e/lib"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"
	"github.com/kyma-project/cloud-manager/pkg/util"
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

func newSimGardenerCluster(kcp client.Client, kubeconfigProvider e2elib.SkrKubeconfigProvider) *simGardenerCluster {
	return &simGardenerCluster{
		kcp:                kcp,
		kubeconfigProvider: kubeconfigProvider,
		clock:              clock.RealClock{},
	}
}

var _ reconcile.Reconciler = &simGardenerCluster{}

type simGardenerCluster struct {
	kcp                client.Client
	kubeconfigProvider e2elib.SkrKubeconfigProvider
	clock              clock.Clock
}

func (r *simGardenerCluster) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	logger := composed.LoggerFromCtx(ctx)
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
		logger.Error(common.ErrLogical, "Missing shoot name label on GardenerCluster")
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

	logger = logger.WithValues("shoot", shootName)

	syncNeeded, expiresIn := e2elib.IsGardenerClusterSyncNeeded(gc, r.clock)
	requeueAfter := expiresIn - 10*time.Second
	if !syncNeeded {
		return reconcile.Result{RequeueAfter: requeueAfter}, nil
	}

	logger.Info("Creating admin kubeconfig for shoot")
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
		logger.Info("Creating kubeconfig secret")
		err = r.kcp.Create(ctx, kubeSecret)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("error creating kubeconfig secret: %w", err)
		}
	} else {
		// found, update it
		kubeSecret.StringData = map[string]string{
			"config": string(kubeConfigBytes),
		}
		logger.Info("Updating kubeconfig secret")
		err = r.kcp.Update(ctx, kubeSecret)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("error updating kubeconfig secret: %w", err)
		}
	}

	if gc.Annotations == nil {
		gc.Annotations = map[string]string{}
	}
	gc.Annotations[e2elib.ExpiresAtAnnotation] = r.clock.Now().Add(r.kubeconfigProvider.ExpiresIn()).Format(time.RFC3339)
	delete(gc.Annotations, e2elib.ForceKubeconfigRotationAnnotation)

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

	return reconcile.Result{RequeueAfter: util.Timing.T300000ms()}, nil
}

func (r *simGardenerCluster) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrastructuremanagerv1.GardenerCluster{}).
		Complete(r)
}
