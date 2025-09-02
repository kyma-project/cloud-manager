package sim

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kyma-project/cloud-manager/api"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/config/crd"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/external/operatorv1beta2"
	"github.com/kyma-project/cloud-manager/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

func NewSimKymaKcp(mgr ctrl.Manager, kcp client.Client, skrProvider SkrProvider) error {
	rec := &simKymaKcp{
		kcp:         kcp,
		skrProvider: skrProvider,
		managers:    map[string]*skrManagerInfo{},
	}
	return rec.SetupWithManager(mgr)
}

type simKymaKcp struct {
	m           sync.Mutex
	kcp         client.Client
	skrProvider SkrProvider
	managers    map[string]*skrManagerInfo
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

	skr, err := r.skrProvider.GetSKR(ctx, kcpKyma.Name)
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("error getting KCP SKR: %w", err)
	}

	skrKyma := &operatorv1beta2.Kyma{}
	err = skr.GetClient().Get(ctx, types.NamespacedName{
		Namespace: "kyma-system",
		Name:      "kyma",
	}, skrKyma)
	if client.IgnoreNotFound(err) != nil && !meta.IsNoMatchError(err) {
		return reconcile.Result{}, fmt.Errorf("error loading SKR Kyma: %w", err)
	}
	if err != nil {
		skrKyma = nil
	}

	// delete ========================================================

	if kcpKyma.DeletionTimestamp != nil {

		r.m.Lock()
		mi, ok := r.managers[kcpKyma.Name]
		if ok && mi != nil {
			logger.Info("Stopping SKR manager")
			mi.cancel()
			mi.wg.Wait()
			if mi.err != nil {
				logger.Error(err, "SKR manager stopped with error")
			} else {
				logger.Info("SKR manager stopped")
			}
		}
		r.m.Unlock()

		// delete skr kyma
		if skrKyma != nil && skrKyma.DeletionTimestamp == nil {
			logger.Info("Deleting SKR Kyma")
			err = skr.GetClient().Delete(ctx, skrKyma)
			if err != nil {
				return reconcile.Result{}, fmt.Errorf("error deleting SKR Kyma: %w", err)
			}
		}

		// wait skr kyma deleted
		if skrKyma != nil {
			logger.Info("Waiting SKR Kyma is deleted...")
			return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
		}

		removed, err := composed.PatchObjRemoveFinalizer(ctx, api.CommonFinalizerDeletionHook, kcpKyma, r.kcp)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("error removing KCP Kyma finalizer: %w", err)
		}
		if removed {
			logger.Info("Removed KCP Kyma Finalizer")
		}

		return reconcile.Result{}, nil
	}

	// create ===========================================================

	// namespace

	ns := &corev1.Namespace{}
	err = skr.GetClient().Get(ctx, types.NamespacedName{Name: "kyma-system"}, ns)
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
		err = skr.GetClient().Create(ctx, ns)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("error creating SKR kyma-system namespace: %w", err)
		}
	}

	// Kyma CRD

	_, err = skr.GetClient().RESTMapper().RESTMapping(operatorv1beta2.GroupVersion.WithKind("Kyma").GroupKind(), operatorv1beta2.GroupVersion.Version)
	if meta.IsNoMatchError(err) {
		arr, err := crd.KLM()
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("error reading Kyma CRD: %w", err)
		}
		logger.Info("Installing Kyma CRD")
		err = util.Apply(ctx, skr.GetClient(), arr)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("error installing Kyma CRD: %w", err)
		}
	}

	// CloudManager module CRD

	_, err = skr.GetClient().RESTMapper().RESTMapping(cloudresourcesv1beta1.GroupVersion.WithKind("CloudResources").GroupKind(), cloudresourcesv1beta1.GroupVersion.Version)
	if meta.IsNoMatchError(err) {
		arr, err := crd.SKR_CloudManagerModule()
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("error reading CloudManager module CRD: %w", err)
		}
		logger.Info("Installing CloudManager module CRD")
		err = util.Apply(ctx, skr.GetClient(), arr)
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
		err = skr.GetClient().Create(ctx, skrKyma)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("error creating SKR kyma: %w", err)
		}
	}

	// manager

	r.m.Lock()
	defer r.m.Unlock()

	_, ok := r.managers[request.Name]
	if !ok {
		mCtx, cancel := context.WithCancel(ctx)
		logger.Info("Starting SKR manager")
		mngr := NewManager(skr, logger)
		err = NewSimKymaSkr(mngr, r.kcp, skr.GetClient())
		if err != nil {
			cancel()
			return reconcile.Result{}, fmt.Errorf("error creating SKR manager: %w", err)
		}
		mi := &skrManagerInfo{
			mngr:   mngr,
			ctx:    mCtx,
			cancel: cancel,
			wg:     &sync.WaitGroup{},
		}
		mi.wg.Add(1)
		r.managers[request.Name] = mi

		go func() {
			err := mngr.Start(mCtx)
			mi.err = err
			mi.wg.Done()
			if err != nil {
				logger.Error(err, "error running manager")
			}
		}()
	}

	return reconcile.Result{}, nil
}

func (r *simKymaKcp) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1beta2.Kyma{}).
		Complete(r)
}
