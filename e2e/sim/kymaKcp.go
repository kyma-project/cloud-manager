package sim

import (
	"context"
	"errors"
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
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
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
	runtimeId        string
	mngr             manager.Manager
	controllersAdded bool
	ctx              context.Context
	cancel           context.CancelFunc
	wg               *sync.WaitGroup
	err              error
}

func newSimKymaKcp(startCtx context.Context, kcp client.Client, clientFactory SkrManagerFactory, kcpNamespace string) *simKymaKcp {
	return &simKymaKcp{
		kcp:                 kcp,
		clientFactory:       clientFactory,
		managersByRuntimeId: map[string]*skrManagerInfo{},
		startCtx:            startCtx,
		kcpNamespace:        kcpNamespace,
	}
}

type simKymaKcp struct {
	m                   sync.Mutex
	kcp                 client.Client
	clientFactory       SkrManagerFactory
	managersByRuntimeId map[string]*skrManagerInfo
	kcpNamespace        string

	// startCtx must be set to the same context that's used to start the sim manager that is
	// running runtime, gardenerCluster and kcpKyma controllers, so also all skrKyma managers are stopped
	// when the startCtx is canceled.
	startCtx context.Context
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

	mi, err := r.getOrCreateStartedSkrManager(ctx, request.Name, logger)
	if errors.Is(err, ErrGardenerClusterCredentialsExpired) {
		// GardenerCluster credentials are invalid (non-existing or expired),
		// have to let the GardenerCluster reconciler create it first, giving it some time to do it
		logger.Info("Out of sync GardenerCluster credentials (expired or not found), waiting for it to be recreated...")
		return reconcile.Result{RequeueAfter: util.Timing.T1000ms()}, nil
	}

	if err != nil {
		logger.Error(err, "Failed to create SKR manager")
		kcpKyma.Status.Conditions = nil
		meta.SetStatusCondition(&kcpKyma.Status.Conditions, metav1.Condition{
			Type:               cloudresourcesv1beta1.ConditionTypeError,
			Status:             metav1.ConditionTrue,
			ObservedGeneration: kcpKyma.Generation,
			Reason:             "ClusterClientCreationError",
			Message:            err.Error(),
		})
		kcpKyma.Status.State = operatorshared.StateError
		_ = composed.PatchObjStatus(ctx, kcpKyma, r.kcp)
		return reconcile.Result{RequeueAfter: util.Timing.T10000ms()}, err
	}

	skrKyma := &operatorv1beta2.Kyma{}
	err = mi.mngr.GetClient().Get(ctx, types.NamespacedName{
		Namespace: "kyma-system",
		Name:      "default",
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
			err = mi.mngr.GetClient().Delete(ctx, skrKyma)
			if err != nil {
				return reconcile.Result{}, fmt.Errorf("error deleting SKR Kyma: %w", err)
			}
		}

		// wait skr kyma deleted

		if skrKyma != nil {
			logger.Info("Waiting SKR Kyma is deleted...")
			return reconcile.Result{RequeueAfter: util.Timing.T10000ms()}, nil
		}

		// stop manager

		logger.Info("Stopping SKR manager")
		mi.cancel()
		mi.wg.Wait()
		if mi.err != nil {
			logger.Error(err, "SKR manager stopped with error")
		} else {
			logger.Info("SKR manager stopped")
		}
		r.m.Lock()
		delete(r.managersByRuntimeId, request.Name)
		r.m.Unlock()

		// remove finalizer

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
		logger.Info("Adding finalizer to KCP Kyma")
		_, err = composed.PatchObjAddFinalizer(ctx, api.CommonFinalizerDeletionHook, kcpKyma, r.kcp)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("error adding KCP Kyma finalizer: %w", err)
		}
	}

	// namespace

	ns := &corev1.Namespace{}
	err = mi.mngr.GetClient().Get(ctx, types.NamespacedName{Name: "kyma-system"}, ns)
	if client.IgnoreNotFound(err) != nil {
		return reconcile.Result{}, fmt.Errorf("error loading SKR kyma-system namespace: %w", err)
	}
	if err != nil {
		ns = nil
	}
	if ns == nil {
		ns = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "kyma-system",
			},
		}
		logger.Info("Creating kyma-system namespace")
		err = mi.mngr.GetClient().Create(ctx, ns)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("error creating SKR kyma-system namespace: %w", err)
		}
	}

	// Kyma CRD

	_, err = mi.mngr.GetClient().RESTMapper().RESTMapping(operatorv1beta2.GroupVersion.WithKind("Kyma").GroupKind(), operatorv1beta2.GroupVersion.Version)
	if meta.IsNoMatchError(err) {
		arr, err := crd.KLM()
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("error reading Kyma CRD: %w", err)
		}
		logger.Info("Installing Kyma CRD")
		err = util.Apply(ctx, mi.mngr.GetClient(), arr)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("error installing Kyma CRD: %w", err)
		}
	}

	// CloudManager module CRD

	_, err = mi.mngr.GetClient().RESTMapper().RESTMapping(cloudresourcesv1beta1.GroupVersion.WithKind("CloudResources").GroupKind(), cloudresourcesv1beta1.GroupVersion.Version)
	if meta.IsNoMatchError(err) {
		arr, err := crd.SKR_CloudManagerModule()
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("error reading CloudManager module CRD: %w", err)
		}
		logger.Info("Installing CloudManager module CRD")
		err = util.Apply(ctx, mi.mngr.GetClient(), arr)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("error installing CloudManager module CRD: %w", err)
		}
	}

	if ns.Status.Phase != corev1.NamespaceActive {
		logger.Info("Waiting for kyma-system namespace to be active")
		return reconcile.Result{RequeueAfter: util.Timing.T1000ms()}, nil
	}

	var ise bool
	isCrdEstablished := func(name string) (bool, error) {
		crdObj := &apiextensionsv1.CustomResourceDefinition{}
		err = mi.mngr.GetClient().Get(ctx, client.ObjectKey{Name: name}, crdObj)
		if err != nil {
			return false, err
		}
		for _, cond := range crdObj.Status.Conditions {
			if cond.Type == apiextensionsv1.Established && cond.Status == apiextensionsv1.ConditionTrue {
				return true, nil
			}
		}
		return false, nil
	}

	ise, err = isCrdEstablished("kymas.operator.kyma-project.io")
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("error checking if kyma CRD is Established: %w", err)
	}
	if !ise {
		logger.Info("Waiting for kyma CRD to be established")
		return reconcile.Result{RequeueAfter: util.Timing.T10000ms()}, nil
	}

	ise, err = isCrdEstablished("cloudresources.cloud-resources.kyma-project.io")
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("error checking if cloudresources CRD is Established: %w", err)
	}
	if !ise {
		logger.Info("Waiting for cloudresources CRD to be established")
		return reconcile.Result{RequeueAfter: util.Timing.T10000ms()}, nil
	}

	// Kyma CR

	if skrKyma == nil {
		skrKyma = &operatorv1beta2.Kyma{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "kyma-system",
				Name:      "default",
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
		err = mi.mngr.GetClient().Create(ctx, skrKyma)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("error creating SKR kyma: %w", err)
		}
	}

	if !mi.controllersAdded {
		err = newSimKymaSkr(r.kcp, mi.mngr.GetClient(), kcpKyma.Name, r.kcpNamespace).
			SetupWithManager(mi.mngr)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("error adding Kyma SKR reconciler to manager: %w", err)
		}
		mi.controllersAdded = true
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
		logger.Info("Patching KCP Kyma status to Ready")
		err = composed.PatchObjStatus(ctx, kcpKyma, r.kcp)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("error patching KCP Kyma status: %w", err)
		}
	}

	return reconcile.Result{}, nil
}

func (r *simKymaKcp) getOrCreateStartedSkrManager(ctx context.Context, runtimeID string, logger logr.Logger) (*skrManagerInfo, error) {
	r.m.Lock()
	defer r.m.Unlock()

	mi, ok := r.managersByRuntimeId[runtimeID]
	if ok {
		return mi, nil
	}

	m, err := r.clientFactory.CreateSkrManager(ctx, runtimeID)
	if err != nil {
		return nil, fmt.Errorf("error creating SKR manager: %w", err)
	}

	mCtx, cancel := context.WithCancel(r.startCtx)
	mCtx = ctrl.LoggerInto(mCtx, ctrl.Log) // clear the logging values
	logger.Info("Starting SKR manager and cluster")

	mi = &skrManagerInfo{
		runtimeId: runtimeID,
		mngr:      m,
		ctx:       mCtx,
		cancel:    cancel,
		wg:        &sync.WaitGroup{},
	}
	mi.wg.Add(1)

	go func() {
		err := m.Start(mCtx)
		mi.err = err
		mi.wg.Done()
		if err != nil {
			logger.Error(err, "error running manager")
		}
	}()

	cacheCtx, cacheCancel := context.WithTimeout(mi.ctx, 20*time.Second)
	defer cacheCancel()
	ok = m.GetCache().WaitForCacheSync(cacheCtx)
	if !ok {
		cancel()
		return nil, fmt.Errorf("error waiting for cache to sync: %w", cacheCtx.Err())
	}

	r.managersByRuntimeId[runtimeID] = mi

	return mi, nil
}

func (r *simKymaKcp) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("kyma-kcp").
		For(&operatorv1beta2.Kyma{}).
		Complete(r)
}
