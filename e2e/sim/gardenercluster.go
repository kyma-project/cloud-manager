package sim

import (
	"context"
	"fmt"
	"time"

	authenticationv1alpha1 "github.com/gardener/gardener/pkg/apis/authentication/v1alpha1"
	gardenertypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	e2econfig "github.com/kyma-project/cloud-manager/e2e/config"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/clock"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	expiresAtAnnotation               = "operator.kyma-project.io/expires-at"
	forceKubeconfigRotationAnnotation = "operator.kyma-project.io/force-kubeconfig-rotation"
	clusterCRNameLabel                = "operator.kyma-project.io/cluster-name"
)

func NewSimGardenerCluster(mgr ctrl.Manager, kcp client.Client, garden client.Client) error {
	rec := &simGardenerCluster{
		kcp:    kcp,
		garden: garden,
	}
	return rec.SetupWithManager(mgr)
}

type simGardenerCluster struct {
	kcp    client.Client
	garden client.Client
	clock clock.Clock
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

	shootName := ""
	if gc.Labels != nil {
		shootName = gc.Labels[clusterCRNameLabel]
	}

	if shootName == "" {
		gc.Status.State = infrastructuremanagerv1.ErrorState
		meta.SetStatusCondition(&gc.Status.Conditions, metav1.Condition{
			Type:               string(infrastructuremanagerv1.ConditionTypeRuntimeKubeconfigReady),
			Status:             metav1.ConditionFalse,
			ObservedGeneration: gc.Generation,
			Reason:             string(infrastructuremanagerv1.ConditionReasonConfigurationErr),
			Message:            fmt.Sprintf("Runtime missing label %s", clusterCRNameLabel),
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

	shoot := &gardenertypes.Shoot{}
	err = r.garden.Get(ctx, types.NamespacedName{
		Namespace: e2econfig.Config.GardenNamespace,
		Name:      shootName,
	}, shoot)
	if err != nil {
		gc.Status.State = infrastructuremanagerv1.ErrorState
		meta.SetStatusCondition(&gc.Status.Conditions, metav1.Condition{
			Type:               string(infrastructuremanagerv1.ConditionTypeRuntimeKubeconfigReady),
			Status:             metav1.ConditionFalse,
			ObservedGeneration: gc.Generation,
			Reason:             string(infrastructuremanagerv1.ConditionReasonGardenerError),
			Message:            fmt.Sprintf("Error getting shoot %s: %s", shootName, err.Error()),
		})
		err = composed.PatchObjStatus(ctx, gc, r.kcp)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("error patching GardenerCluster status with error getting shoot condition: %w", err)
		}
		return reconcile.Result{}, nil
	}

	expiresIn := 6 * time.Hour

	adminKubeconfigRequest := &authenticationv1alpha1.AdminKubeconfigRequest{
		Spec: authenticationv1alpha1.AdminKubeconfigRequestSpec{
			ExpirationSeconds: ptr.To(int64(expiresIn.Seconds())),
		},
	}
	err = r.garden.SubResource("adminkubeconfig").Create(ctx, shoot, adminKubeconfigRequest)
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("error creating admin kubeconfig: %w", err)
	}
	kubeConfigBytes := adminKubeconfigRequest.Status.Kubeconfig

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
	gc.Annotations[expiresAtAnnotation] = r.clock.Now().Add(expiresIn).Format(time.RFC3339)
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
	if gc.Annotations == nil {
		return true, time.Minute
	}
	if _, ok := gc.Annotations[forceKubeconfigRotationAnnotation]; !ok {
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

	return true, time.Duration(0.5 * float64(expiresIn))
}

func (r *simGardenerCluster) getShootKubeconfig(ctx context.Context, shoot *gardenertypes.Shoot) ([]byte, error) {
	adminKubeconfigRequest := &authenticationv1alpha1.AdminKubeconfigRequest{
		Spec: authenticationv1alpha1.AdminKubeconfigRequestSpec{
			ExpirationSeconds: ptr.To(int64(3600 * 6)), // 6 hours
		},
	}
	err := r.garden.SubResource("adminkubeconfig").Create(ctx, shoot, adminKubeconfigRequest)
	if err != nil {
		return nil, fmt.Errorf("error creating admin kubeconfig: %w", err)
	}
	return adminKubeconfigRequest.Status.Kubeconfig, nil
}

func (r *simGardenerCluster) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrastructuremanagerv1.GardenerCluster{}).
		Complete(r)
}
