package mock2

import (
	"context"
	"testing"
	"time"

	gcputil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/util"
	"github.com/stretchr/testify/assert"
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

		rc := s.createRedisClusterOK(parentNd.String(), net.GetSelfLink(), "test-cluster", 3, 4, nil)

		rcName, err := gcputil.ParseNameDetail(rc.Name)
		assert.NoError(s.t, err)
		assert.Equal(t, gcputil.ResourceTypeCluster, rcName.ResourceType())

		assert.EqualValues(t, 3, rc.GetReplicaCount())
		assert.EqualValues(t, 4, rc.GetShardCount())

		// delete

		s.deleteRedisClusterOK(rc.GetName())
		s.deleteServiceConnectionPolicyOK(scp.Name)
		s.deleteSubnetOK(sub.GetRegion(), sub.GetName())
		s.deleteNetworkOK(net.GetName())
	})
}
