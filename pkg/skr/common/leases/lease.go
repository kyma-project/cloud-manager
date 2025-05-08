package leases

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/api"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/skr/runtime/config"
	"k8s.io/api/coordination/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sync"
	"time"
)

type LeaseResult int

const (
	RenewedLease  LeaseResult = iota
	AcquiredLease LeaseResult = iota
	LeasingFailed LeaseResult = iota
	OtherLeased   LeaseResult = iota
)

var leaseMutex sync.Mutex

func Acquire(ctx context.Context, cluster composed.StateCluster, resourceName, ownerName types.NamespacedName, prefix string) (LeaseResult, error) {
	leaseMutex.Lock()
	defer leaseMutex.Unlock()

	leaseName := getLeaseName(resourceName, prefix)
	leaseNamespace := resourceName.Namespace
	holderName := getHolderName(ownerName)

	lease := &v1.Lease{}
	err := cluster.K8sClient().Get(ctx, types.NamespacedName{Name: leaseName, Namespace: leaseNamespace}, lease)
	if err != nil && !apierrors.IsNotFound(err) {
		return LeasingFailed, err
	}
	leaseDuration := int32(config.SkrRuntimeConfig.SkrLockingLeaseDuration.Seconds())
	if lease.Name == "" {
		lease = &v1.Lease{
			ObjectMeta: metav1.ObjectMeta{
				Name:      leaseName,
				Namespace: leaseNamespace,
			},
			Spec: v1.LeaseSpec{
				HolderIdentity:       &holderName,
				LeaseDurationSeconds: &leaseDuration,
				AcquireTime:          &metav1.MicroTime{Time: time.Now()},
				RenewTime:            &metav1.MicroTime{Time: time.Now()},
			},
		}
		controllerutil.AddFinalizer(lease, api.CommonFinalizerDeletionHook)
		err := cluster.K8sClient().Create(ctx, lease)
		if err != nil {
			return LeasingFailed, err
		}
		return AcquiredLease, nil
	}
	if ptr.Deref(lease.Spec.HolderIdentity, "") == holderName {
		lease.Spec.RenewTime = &metav1.MicroTime{Time: time.Now()}
		err := cluster.K8sClient().Update(ctx, lease)
		if err != nil {
			return LeasingFailed, err
		}
		return RenewedLease, nil
	}
	leaseDurationSeconds := ptr.Deref(lease.Spec.LeaseDurationSeconds, leaseDuration)
	if lease.Spec.RenewTime == nil || lease.Spec.RenewTime.Time.Add(time.Second*time.Duration(leaseDurationSeconds)).Before(time.Now()) {
		//lease expired
		lease.Spec.RenewTime = &metav1.MicroTime{Time: time.Now()}
		lease.Spec.HolderIdentity = &holderName
		err := cluster.K8sClient().Update(ctx, lease)
		if err != nil {
			return LeasingFailed, err
		}
		return AcquiredLease, nil
	}
	return OtherLeased, nil
}

func Release(ctx context.Context, cluster composed.StateCluster, resourceName, ownerName types.NamespacedName, prefix string) error {
	leaseMutex.Lock()
	defer leaseMutex.Unlock()
	leaseName := getLeaseName(resourceName, prefix)
	leaseNamespace := resourceName.Namespace
	holderName := getHolderName(ownerName)
	lease := &v1.Lease{}
	err := cluster.K8sClient().Get(ctx, types.NamespacedName{Name: leaseName, Namespace: leaseNamespace}, lease)
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	if lease.Name != "" && ptr.Deref(lease.Spec.HolderIdentity, "") == holderName {
		controllerutil.RemoveFinalizer(lease, api.CommonFinalizerDeletionHook)
		// patch is preferable to update, but fake client does not support it causing tests to fail
		err = cluster.K8sClient().Update(ctx, lease)
		if err != nil {
			return err
		}
		// getting the lease again to avoid revision conflict
		err = cluster.K8sClient().Get(ctx, types.NamespacedName{Name: leaseName, Namespace: leaseNamespace}, lease)
		if err != nil && !apierrors.IsNotFound(err) {
			return err
		}
		err = cluster.K8sClient().Delete(ctx, lease)
		if err != nil {
			return err
		}
	}
	return nil
}

func getHolderName(ownerName types.NamespacedName) string {
	return fmt.Sprintf("%s/%s", ownerName.Namespace, ownerName.Name)
}

func getLeaseName(resourceName types.NamespacedName, prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, resourceName.Name)
}
