package leases

import (
	"context"
	"fmt"
	"time"

	"github.com/kyma-project/cloud-manager/api"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	v1 "k8s.io/api/coordination/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type LeaseResult int

const (
	RenewedLease LeaseResult = iota
	AcquiredLease
	LeasingFailed
	OtherLeased
)

func Acquire(ctx context.Context, cluster composed.StateCluster, leaseName, leaseNamespace, holderName string, leaseDurationSec int32) (LeaseResult, error) {
	lease := &v1.Lease{}
	err := cluster.K8sClient().Get(ctx, types.NamespacedName{Name: leaseName, Namespace: leaseNamespace}, lease)
	if err != nil && !apierrors.IsNotFound(err) {
		return LeasingFailed, err
	}

	if lease.Name == "" {
		lease = &v1.Lease{
			ObjectMeta: metav1.ObjectMeta{
				Name:      leaseName,
				Namespace: leaseNamespace,
			},
			Spec: v1.LeaseSpec{
				HolderIdentity:       &holderName,
				LeaseDurationSeconds: &leaseDurationSec,
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
	leaseDurationSeconds := ptr.Deref(lease.Spec.LeaseDurationSeconds, leaseDurationSec)
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

func Release(ctx context.Context, cluster composed.StateCluster, leaseName, leaseNamespace, holderName string) error {
	lease := &v1.Lease{}
	err := cluster.K8sClient().Get(ctx, types.NamespacedName{Name: leaseName, Namespace: leaseNamespace}, lease)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	ownerName := ptr.Deref(lease.Spec.HolderIdentity, "")

	if ownerName != holderName {
		return fmt.Errorf("lease %s/%s belongs to another owner (%s)", leaseName, leaseNamespace, holderName)
	}

	controllerutil.RemoveFinalizer(lease, api.CommonFinalizerDeletionHook)
	err = cluster.K8sClient().Update(ctx, lease)
	if err != nil {
		return err
	}
	err = cluster.K8sClient().Delete(ctx, lease)
	if err != nil {
		return err
	}
	return nil
}
