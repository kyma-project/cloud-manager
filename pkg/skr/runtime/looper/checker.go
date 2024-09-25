package looper

import (
	"context"
	"github.com/go-logr/logr"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"time"
)

type Checker interface {
	IsReady(ctx context.Context, skrCluster cluster.Cluster) bool
}

var _ Checker = &checker{}

type checker struct {
	logger logr.Logger
}

func (c *checker) IsReady(ctx context.Context, skrCluster cluster.Cluster) bool {
	timeout := time.Now().Add(time.Millisecond * 3500)
	interval := time.Second
	for {
		select {
		case <-ctx.Done():
			c.logger.Info("SKR Checker context closed")
			return false
		default:
			ok := c.singleCheck(ctx, skrCluster)
			if ok {
				return true
			}
			if timeout.Before(time.Now()) {
				c.logger.Info("Timeout checking SKR readiness")
				return false
			}
			time.Sleep(interval)
		}
	}
}

// singleCheck returns true if it is ok to proceed with SKR connection that implies installation of CRDs and
// running controllers. It is important to consider both provisioning and deprovisioning phase:
// * Provisioning
//   - CloudResources CRD is not yet installed by KLM
//   - CloudResources default instance is not yet created by KLM
//
// * Deprovisioning
//   - There are no CloudResources instances since they all have been deleted - KLM marked for deletion,
//     CloudManager connected to SKR and uninstalled CRDs and removed finalized, K8S deleted CloudResources.
//     If CloudManager connects again before the KCP Scope is reconciled and this SKR removed from active
//     we must ensure connection does not proceed and installs CRDs again.
//   - There is CloudResources instance, but it's marked for deletion and has no finalizer since CloudManager
//     already have connected, deleted CRDs and removed finalizer, and now K8S API should delete this resource.
//     Not sure if practically possible, but implementing it just in case.
func (c *checker) singleCheck(ctx context.Context, skrCluster cluster.Cluster) bool {
	list := &cloudresourcesv1beta1.CloudResourcesList{}
	err := skrCluster.GetAPIReader().List(ctx, list)
	if err != nil {
		c.logger.Error(err, "SKR readiness failed - CloudResources CRD not installed")
		return false
	}
	if len(list.Items) == 0 {
		c.logger.Error(err, "SKR readiness failed - no CloudResources created")
		return false
	}
	allDeletedAndNoFinalizer := true
	for _, item := range list.Items {
		// has deletion timestamp and has no finalizers
		if !item.DeletionTimestamp.IsZero() && len(item.Finalizers) == 0 {
			continue
		}
		allDeletedAndNoFinalizer = false
		break
	}
	if allDeletedAndNoFinalizer {
		c.logger.Error(err, "SKR readiness failed - all CloudResources instances are being deleted and have no finalizer")
		return false
	}
	return true
}
