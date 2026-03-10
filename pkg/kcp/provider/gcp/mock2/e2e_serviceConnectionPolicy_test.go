package mock2

import (
	"context"
	"testing"
	"time"

	gcputil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/util"
	"github.com/stretchr/testify/require"
)

func TestE2EServiceConnectionPolicy(t *testing.T) {

	t.Run("ServiceConnectionPolicy is created with one subnets, updated with second subnet and deleted", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		s := newE2ETestSuite(ctx, t)
		region := "us-east1"
		subnetId1 := "test-subnet-1"
		subnetId2 := "test-subnet-2"

		net := s.createNetworkOK("test-net")

		parentNd := gcputil.NewLocationName(s.mock.ProjectId(), region)

		// subnet 1
		sub1 := s.createSubnetOK(region, net.GetSelfLink(), subnetId1, "10.250.0.0/16")
		scp := s.createServiceConnectionPolicyOK(parentNd.String(), "test-policy", net.GetSelfLink(), []string{sub1.GetSelfLink()})

		// subnet 2
		sub2 := s.createSubnetOK(region, net.GetSelfLink(), subnetId2, "10.251.0.0/16")
		scp.PscConfig.Subnetworks = append(scp.PscConfig.Subnetworks, sub2.GetSelfLink())

		scp = s.updateServiceConnectionPolicyOK(scp, []string{"psc_config"})
		require.Len(t, scp.PscConfig.Subnetworks, 2)
		require.Equal(t, sub1.GetSelfLink(), scp.PscConfig.Subnetworks[0])
		require.Equal(t, sub2.GetSelfLink(), scp.PscConfig.Subnetworks[1])

		// delete
		s.deleteServiceConnectionPolicyOK(scp.Name)
		s.deleteSubnetOK(sub1.GetRegion(), sub1.GetName())
		s.deleteSubnetOK(sub2.GetRegion(), sub2.GetName())
		s.deleteNetworkOK(net.GetName())
	})

	t.Run("Subnet can not be deleted when used by serviceConnectionPolicy", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		s := newE2ETestSuite(ctx, t)
		region := "us-east1"
		net := s.createNetworkOK("test-net")
		parentNd := gcputil.NewLocationName(s.mock.ProjectId(), region)
		sub1 := s.createSubnetOK(region, net.GetSelfLink(), "test-subnet", "10.250.0.0/16")
		scp := s.createServiceConnectionPolicyOK(parentNd.String(), "test-policy", net.GetSelfLink(), []string{sub1.GetSelfLink()})

		op, err := s.deleteSubnet(sub1.GetRegion(), sub1.GetName())
		require.Error(t, err)
		require.Nil(t, op)

		s.deleteServiceConnectionPolicyOK(scp.Name)
		s.deleteSubnetOK(sub1.GetRegion(), sub1.GetName())
	})
}
