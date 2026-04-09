package mock2

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestE2EServiceNetworkingConnection(t *testing.T) {

	t.Run("SNC PSA connection can be created and deleted", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		s := newE2ETestSuite(ctx, t)

		net := s.createNetworkOK("test-net")

		// first address
		addr := s.createPsaRangeOK(net.GetSelfLink(), "test-address", "10.251.0.0", 16)

		con := s.createPsaConnectionOK(net.GetSelfLink(), addr.GetSelfLink())
		require.Len(t, con.ReservedPeeringRanges, 1)
		require.Equal(t, []string{addr.GetName()}, con.ReservedPeeringRanges)

		// second address
		addr2 := s.createPsaRangeOK(net.GetSelfLink(), "test-address-2", "10.252.0.0", 16)

		op, err := s.mock.PatchServiceConnection(s.ctx, s.mock.ProjectId(), net.GetName(), []string{addr.GetName(), addr2.GetName()})
		require.NoError(t, err)
		require.True(t, op.Done)

		arr, err := s.mock.ListServiceConnections(s.ctx, s.mock.ProjectId(), net.GetName())
		require.NoError(t, err)
		require.Len(t, arr, 1)
		require.Len(t, arr[0].ReservedPeeringRanges, 2)
		require.Equal(t, []string{addr.GetName(), addr2.GetName()}, arr[0].ReservedPeeringRanges)

		// delete

		s.deletePsaConnectionOK(net.GetSelfLink())
		s.deleteAddressOK(addr.GetName())
		s.deleteAddressOK(addr2.GetName())
		s.deleteNetworkOK(net.GetName())
	})
}
