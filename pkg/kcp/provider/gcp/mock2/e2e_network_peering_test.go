package mock2

import (
	"context"
	"testing"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2ENetworkPeering(t *testing.T) {

	t.Run("peer two networks", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		srv := New()

		storeLocal := &e2eTestSuite{
			ctx:  ctx,
			t:    t,
			mock: srv.NewSubscription("local-"),
		}
		storeRemote := &e2eTestSuite{
			ctx:  ctx,
			t:    t,
			mock: srv.NewSubscription("remote-"),
		}

		netLocal := storeLocal.createNetworkOK("local-net")
		netRemote := storeRemote.createNetworkOK("remote-net")

		peeringLocal := storeLocal.addPeeringOK(netLocal.GetSelfLink(), "local-peering", netRemote.GetSelfLink())
		assert.Equal(t, computepb.NetworkPeering_INACTIVE.String(), peeringLocal.GetState())

		peeringRemote := storeRemote.addPeeringOK(netRemote.GetSelfLink(), "remote-peering", netLocal.GetSelfLink())
		assert.Equal(t, computepb.NetworkPeering_ACTIVE.String(), peeringRemote.GetState())

		// check if local peering is ACTIVE
		peeringLocal = storeLocal.getPeering(netLocal.GetSelfLink(), peeringLocal.GetName())
		require.NotNil(t, peeringLocal)
		assert.Equal(t, computepb.NetworkPeering_ACTIVE.String(), peeringLocal.GetState())

		// delete local peering

		storeLocal.removePeeringOK(netLocal.GetSelfLink(), peeringLocal.GetName())

		// check if remote peering is INACTIVE
		peeringRemote = storeRemote.getPeering(netRemote.GetSelfLink(), peeringRemote.GetName())
		require.NotNil(t, peeringRemote)
		assert.Equal(t, computepb.NetworkPeering_INACTIVE.String(), peeringRemote.GetState())

		// delete remote peering
		storeRemote.removePeeringOK(netRemote.GetSelfLink(), peeringRemote.GetName())

		storeLocal.deleteNetworkOK(netLocal.GetName())
		storeRemote.deleteNetworkOK(netRemote.GetName())
	})

}
