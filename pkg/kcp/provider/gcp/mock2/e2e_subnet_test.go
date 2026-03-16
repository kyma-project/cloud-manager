package mock2

import (
	"context"
	"fmt"
	"testing"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"
	gcputil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2ESubnet(t *testing.T) {
	t.Run("Subnet can be created and deleted", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		s := newE2ETestSuite(ctx, t)

		// Create network first
		net := s.createNetworkOK("test-net")

		// Create subnet in the network
		sub := s.createSubnetOK("us-east1", net.GetSelfLink(), "test-subnet", "10.250.0.0/16")

		// Verify subnet fields are set correctly
		require.NotNil(s.t, sub.Id)
		require.NotZero(s.t, *sub.Id)
		require.Equal(s.t, computepb.Subnetwork_READY.String(), sub.GetState())
		require.Equal(s.t, "compute#subnetwork", sub.GetKind())
		require.Equal(s.t, sub.GetNetwork(), net.GetSelfLink())

		subnetName, err := gcputil.ParseNameDetail(sub.GetSelfLink())
		assert.NoError(s.t, err)
		assert.Equal(t, gcputil.ResourceTypeSubnetwork, subnetName.ResourceType())

		// Clean up resources in reverse order
		s.deleteSubnetOK(subnetName.LocationRegionId(), subnetName.ResourceId())
		s.deleteNetworkOK(net.GetName())
	})

	t.Run("Subnet can not be deleted when used by SCP", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		s := newE2ETestSuite(ctx, t)

		net := s.createNetworkOK("test-net")
		sub := s.createSubnetOK("us-east1", net.GetSelfLink(), "test-subnet", "10.250.0.0/16")

		parentNd := gcputil.NewLocationName(s.mock.ProjectId(), "us-east1")
		_ = s.createServiceConnectionPolicyOK(parentNd.String(), "test-policy", net.GetSelfLink(), []string{sub.GetSelfLink()})

		opVoid, err := s.deleteSubnet(sub.GetRegion(), sub.GetName())
		require.Error(t, err)
		require.Nil(t, opVoid)
		require.Contains(t, err.Error(), fmt.Sprintf("subnet projects/%[1]s/regions/us-east1/subnetworks/test-subnet is attached to SCP projects/%[1]s/locations/us-east1/serviceConnectionPolicies/test-policy", s.mock.ProjectId()))
	})
}
