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
	timeout := time.Now().Add(time.Second * 3)
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
			}
			time.Sleep(interval)
		}
	}
}

func (c *checker) singleCheck(ctx context.Context, skrCluster cluster.Cluster) bool {
	list := &cloudresourcesv1beta1.CloudResourcesList{}
	err := skrCluster.GetAPIReader().List(ctx, list)
	if err != nil {
		c.logger.Error(err, "SKR readiness failed")
		return false
	}
	return true
}
