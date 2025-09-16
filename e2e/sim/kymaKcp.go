package sim

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/api"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/config/crd"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/external/operatorshared"
	"github.com/kyma-project/cloud-manager/pkg/external/operatorv1beta2"
	"github.com/kyma-project/cloud-manager/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type skrManagerInfo struct {
	mngr   manager.Manager
	ctx    context.Context
	cancel context.CancelFunc
	wg     *sync.WaitGroup
	err    error
}

func newSimKymaKcp(kcp client.Client, clientFactory ClientClusterFactory) *simKymaKcp {
	return &simKymaKcp{
		kcp:           kcp,
		clientFactory: clientFactory,
		managers:      map[string]*skrManagerInfo{},
	}
}

type simKymaKcp struct {
	m             sync.Mutex
	kcp           client.Client
	clientFactory ClientClusterFactory
	managers      map[string]*skrManagerInfo
}

func (r *simKymaKcp) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	logger := composed.LoggerFromCtx(ctx)
	kcpKyma := &operatorv1beta2.Kyma{}
	err := r.kcp.Get(ctx, request.NamespacedName, kcpKyma)
	if client.IgnoreNotFound(err) != nil {
		return reconcile.Result{}, fmt.Errorf("error loading KCP Kyma: %w", err)
	}
	if err != nil {
		return reconcile.Result{}, nil
	}

	if _, ok := kcpKyma.Labels[DoNotReconcile]; ok {
		return reconcile.Result{}, nil
	}

	mi, err := r.getManagerInfo(ctx, request.Name, logger)
	if err != nil {
		kcpKyma.Status.State = operatorshared.StateError
		_ = composed.PatchObjStatus(ctx, kcpKyma, r.kcp)
		return reconcile.Result{RequeueAfter: time.Minute}, err
	}

	skrClient := mi.mngr.GetClient()

	skrKyma := &operatorv1beta2.Kyma{}
	err = skrClient.Get(ctx, types.NamespacedName{
		Namespace: "kyma-system",
		Name:      "kyma",
	}, skrKyma)
	if client.IgnoreNotFound(err) != nil && !meta.IsNoMatchError(err) {
		return reconcile.Result{}, fmt.Errorf("error loading SKR Kyma: %w", err)
	}
	if err != nil {
		skrKyma = nil
	}

	// delete =======================================================================

	if kcpKyma.DeletionTimestamp != nil {

		// delete skr kyma
		if skrKyma != nil && skrKyma.DeletionTimestamp == nil {
			logger.Info("Deleting SKR Kyma")
			err = skrClient.Delete(ctx, skrKyma)
			if err != nil {
				return reconcile.Result{}, fmt.Errorf("error deleting SKR Kyma: %w", err)
			}
		}

		// wait skr kyma deleted
		if skrKyma != nil {
			logger.Info("Waiting SKR Kyma is deleted...")
			return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
		}

		r.m.Lock()
		logger.Info("Stopping SKR manager")
		mi.cancel()
		mi.wg.Wait()
		if mi.err != nil {
			logger.Error(err, "SKR manager stopped with error")
		} else {
			logger.Info("SKR manager stopped")
		}
		r.m.Unlock()

		removed, err := composed.PatchObjRemoveFinalizer(ctx, api.CommonFinalizerDeletionHook, kcpKyma, r.kcp)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("error removing KCP Kyma finalizer: %w", err)
		}
		if removed {
			logger.Info("Removed KCP Kyma Finalizer")
		}

		return reconcile.Result{}, nil
	}

	// create ======================================================================

	// finalizer
	if !controllerutil.ContainsFinalizer(kcpKyma, api.CommonFinalizerDeletionHook) {
		_, err = composed.PatchObjAddFinalizer(ctx, api.CommonFinalizerDeletionHook, kcpKyma, r.kcp)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("error adding KCP Kyma finalizer: %w", err)
		}
	}

	// namespace

	ns := &corev1.Namespace{}
	err = skrClient.Get(ctx, types.NamespacedName{Name: "kyma-system"}, ns)
	if client.IgnoreNotFound(err) != nil {
		return reconcile.Result{}, fmt.Errorf("error loading SKR kyma-system namespace: %w", err)
	}
	if err != nil {
		ns = nil
	}
	if ns == nil {
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "kyma-system",
			},
		}
		logger.Info("Creating kyma-system namespace")
		err = skrClient.Create(ctx, ns)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("error creating SKR kyma-system namespace: %w", err)
		}
	}

	// Kyma CRD

	_, err = skrClient.RESTMapper().RESTMapping(operatorv1beta2.GroupVersion.WithKind("Kyma").GroupKind(), operatorv1beta2.GroupVersion.Version)
	if meta.IsNoMatchError(err) {
		arr, err := crd.KLM()
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("error reading Kyma CRD: %w", err)
		}
		logger.Info("Installing Kyma CRD")
		err = util.Apply(ctx, skrClient, arr)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("error installing Kyma CRD: %w", err)
		}
	}

	// CloudManager module CRD

	_, err = skrClient.RESTMapper().RESTMapping(cloudresourcesv1beta1.GroupVersion.WithKind("CloudResources").GroupKind(), cloudresourcesv1beta1.GroupVersion.Version)
	if meta.IsNoMatchError(err) {
		arr, err := crd.SKR_CloudManagerModule()
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("error reading CloudManager module CRD: %w", err)
		}
		logger.Info("Installing CloudManager module CRD")
		err = util.Apply(ctx, skrClient, arr)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("error installing CloudManager module CRD: %w", err)
		}
	}

	// Kyma CR

	if skrKyma == nil {
		skrKyma = &operatorv1beta2.Kyma{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "kyma-system",
				Name:      "kyma",
				Labels:    kcpKyma.Labels,
				Finalizers: []string{
					api.CommonFinalizerDeletionHook,
				},
			},
			Spec: operatorv1beta2.KymaSpec{
				Channel: operatorv1beta2.DefaultChannel,
			},
		}
		logger.Info("Creating SKR Kyma")
		err = skrClient.Create(ctx, skrKyma)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("error creating SKR kyma: %w", err)
		}
	}

	statusChanged := false
	if kcpKyma.Status.State != operatorshared.StateReady {
		kcpKyma.Status.State = operatorshared.StateReady
		statusChanged = true
	}
	if len(kcpKyma.Status.Conditions) > 0 {
		kcpKyma.Status.Conditions = []metav1.Condition{}
		statusChanged = true
	}
	if statusChanged {
		err = composed.PatchObjStatus(ctx, kcpKyma, r.kcp)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("error patching KCP Kyma status: %w", err)
		}
	}

	return reconcile.Result{}, nil
}

func (r *simKymaKcp) getManagerInfo(ctx context.Context, runtimeID string, logger logr.Logger) (*skrManagerInfo, error) {
	r.m.Lock()
	defer r.m.Unlock()

	mi, ok := r.managers[runtimeID]
	if ok {
		return mi, nil
	}

	skrCluster, err := r.clientFactory.CreateClientCluster(ctx, runtimeID)
	if err != nil {
		return nil, fmt.Errorf("error creating SKR Cluster: %w", err)
	}

	mCtx, cancel := context.WithCancel(ctx)
	logger.Info("Starting SKR manager and cluster")

	mngr := NewManager(skrCluster, logger)
	err = NewSimKymaSkr(r.kcp, skrCluster.GetClient()).SetupWithManager(mngr)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("error creating SKR manager: %w", err)
	}
	mi = &skrManagerInfo{
		mngr:   mngr,
		ctx:    mCtx,
		cancel: cancel,
		wg:     &sync.WaitGroup{},
	}
	mi.wg.Add(1)

	go func() {
		err := mngr.Start(mCtx)
		mi.err = err
		mi.wg.Done()
		if err != nil {
			logger.Error(err, "error running manager")
		}
	}()

	cacheCtx, cacheCancel := context.WithTimeout(mi.ctx, 10*time.Minute)
	defer cacheCancel()
	ok = mngr.GetCache().WaitForCacheSync(cacheCtx)
	if !ok {
		cancel()
		return nil, fmt.Errorf("error waiting for cache to sync: %w", cacheCtx.Err())
	}

	r.managers[runtimeID] = mi

	return mi, nil
}

func (r *simKymaKcp) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("kyma-kcp").
		For(&operatorv1beta2.Kyma{}).
		Complete(r)
}
