package mock2

import (
	"context"
	"fmt"
	"testing"
	"time"

	"cloud.google.com/go/filestore/apiv1/filestorepb"
	gcputil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/util"
	"github.com/stretchr/testify/require"
)

func TestE2EFilestore(t *testing.T) {

	t.Run("Filestore instance can be created, updated and deleted", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		s := newE2ETestSuite(ctx, t)

		location := "us-east1"

		// create network

		net := s.createNetworkOK("test-net")

		// create PSA range

		addr := s.createPsaRangeOK(net.GetSelfLink(), "test-address", "10.251.0.0", 16)

		// crete instance

		fs := s.createFilestoreOK(gcputil.NewLocationName(s.mock.ProjectId(), location).String(), "test-instance", net.GetSelfLink(), addr.GetSelfLink(), 1024)

		// assert create fs instance properties

		require.Equal(t, gcputil.NewInstanceName(s.mock.ProjectId(), location, "test-instance").String(), fs.GetName())
		require.Len(t, fs.Networks, 1)
		require.Len(t, fs.Networks[0].IpAddresses, 1)
		require.Equal(t, "10.251.0.0", fs.Networks[0].IpAddresses[0])

		// update instance

		fs = s.updateFilestoreOK(
			&filestorepb.Instance{
				Name: fs.Name,
				FileShares: []*filestorepb.FileShareConfig{
					{
						CapacityGb: 2048,
					},
				},
			},
			[]string{"file_shares"},
		)

		// delete fs instance

		s.deleteFilestoreOK(fs.Name)
	})

	t.Run("PSA address range can not be deleted if used by filestore", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		s := newE2ETestSuite(ctx, t)
		net := s.createNetworkOK("test-net")
		addr := s.createPsaRangeOK(net.GetSelfLink(), "test-address", "10.251.0.0", 16)
		_ = s.createFilestoreOK(gcputil.NewLocationName(s.mock.ProjectId(), "us-east1").String(), "test-instance", net.GetSelfLink(), addr.GetSelfLink(), 1024)

		opVoid, err := s.deleteAddress(addr.GetName())
		require.Error(t, err)
		require.Contains(s.t, err.Error(), fmt.Sprintf("address projects/%[1]s/address/test-address is in use by filestore projects/%[1]s/locations/us-east1/instances/test-instance", s.mock.ProjectId()))
		require.Nil(t, opVoid)
	})

}
