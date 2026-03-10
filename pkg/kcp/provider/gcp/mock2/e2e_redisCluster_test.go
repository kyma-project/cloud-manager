package mock2

import (
	"context"
	"testing"
	"time"

	gcputil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/util"
)

func TestE2ERedisCluster(t *testing.T) {

	t.Run("Redis Cluster can be created, updated and deleted", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		s := newE2ETestSuite(ctx, t)
		region := "us-east1"
		net := s.createNetworkOK("test-net")
		sub := s.createSubnetOK(region, net.GetName(), "test-subnet", "10.250.0.0/16")

		parentNd := gcputil.NewLocationName(s.mock.ProjectId(), region)
		scp := s.createServiceConnectionPolicyOK(parentNd.String(), "redis-cluster", net.GetSelfLink(), []string{sub.GetSelfLink()})

		// delete

		s.deleteServiceConnectionPolicyOK(scp.Name)
		s.deleteSubnetOK(sub.GetRegion(), sub.GetName())
		s.deleteNetworkOK(net.GetName())
	})
}
