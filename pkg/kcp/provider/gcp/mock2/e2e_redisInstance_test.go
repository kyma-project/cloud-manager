package mock2

import (
	"context"
	"fmt"
	"testing"
	"time"

	"cloud.google.com/go/redis/apiv1/redispb"
	gcputil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/util"
	"github.com/stretchr/testify/require"
)

func TestE2ERedisInstance(t *testing.T) {

	t.Run("Redis Instance can be created, updated and deleted", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		s := newE2ETestSuite(ctx, t)

		location := "us-east1"

		// create network

		net := s.createNetworkOK("test-net")

		// create PSA range

		addr := s.createPsaRangeOK(net.GetSelfLink(), "test-address", "10.251.0.0", 16)

		// crete instance

		ri := s.createRedisInstanceOK(gcputil.NewLocationName(s.mock.ProjectId(), location).String(), "test-instance", net.GetSelfLink(), addr.GetSelfLink(), 1024)

		// assert create fs instance properties

		require.Equal(t, gcputil.NewInstanceName(s.mock.ProjectId(), location, "test-instance").String(), ri.GetName())
		require.Equal(t, "10.251.0.0", ri.Host)
	require.EqualValues(t, 6379, ri.Port)

		// update instance

		ri = s.updateRedisInstanceOK(
			&redispb.Instance{
				Name:         ri.Name,
				MemorySizeGb: 2,
				DisplayName:  "something else",
			},
			[]string{"memory_size_gb"},
		)
		// display name was not change since it was not in the updateMask
		require.Equal(t, gcputil.MustParseNameDetail(ri.Name).ResourceId(), ri.DisplayName)

		// delete fs instance

		s.deleteRedisInstanceOK(ri.Name)
	})

	t.Run("PSA address range can not be deleted if used by redisInstance", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		s := newE2ETestSuite(ctx, t)
		net := s.createNetworkOK("test-net")
		addr := s.createPsaRangeOK(net.GetSelfLink(), "test-address", "10.251.0.0", 16)
		_ = s.createRedisInstanceOK(gcputil.NewLocationName(s.mock.ProjectId(), "us-east1").String(), "test-instance", net.GetSelfLink(), addr.GetSelfLink(), 1)

		opVoid, err := s.deleteAddress(addr.GetName())
		require.Error(t, err)
		require.Contains(s.t, err.Error(), fmt.Sprintf("address projects/%[1]s/address/test-address is in use by redisInstance projects/%[1]s/locations/us-east1/instances/test-instance", s.mock.ProjectId()))
		require.Nil(t, opVoid)
	})

}
