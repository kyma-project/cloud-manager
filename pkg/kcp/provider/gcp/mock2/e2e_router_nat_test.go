package mock2

import (
	"context"
	"testing"
	"time"

	gcputil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2ERouterNat(t *testing.T) {

	t.Run("Router and NAT are created and deleted", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		s := newE2ETestSuite(ctx, t)
		region := "us-east1"
		subnetName := "test-subnet"

		net := s.createNetworkOK("test-net")
		sub := s.createSubnetOK(region, net.GetSelfLink(), subnetName, "10.250.0.0/16")
		rt := s.createRouterOK(region, net.GetSelfLink(), "my-router", []string{sub.GetSelfLink()})

		rtName, err := gcputil.ParseNameDetail(rt.GetSelfLink())
		assert.NoError(s.t, err)
		assert.Equal(t, gcputil.ResourceTypeRouter, rtName.ResourceType())

		require.Len(s.t, rt.Nats, 1)
		require.Len(s.t, rt.Nats[0].Subnetworks, 1)
		require.Equal(t, sub.GetSelfLink(), rt.Nats[0].Subnetworks[0].GetName())

		s.deleteRouterOK(region, rt.GetName())
		s.deleteSubnetOK(region, subnetName)
		s.deleteNetworkOK(net.GetName())
	})
}
