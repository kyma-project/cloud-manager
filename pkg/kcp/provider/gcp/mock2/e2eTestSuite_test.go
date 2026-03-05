package mock2

import (
	"context"
	"testing"

	"cloud.google.com/go/compute/apiv1/computepb"
	"cloud.google.com/go/filestore/apiv1/filestorepb"
	"cloud.google.com/go/redis/apiv1/redispb"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
	gcputil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/util"
	"github.com/stretchr/testify/require"
	"google.golang.org/genproto/protobuf/field_mask"
	"k8s.io/utils/ptr"
)

type e2eTestSuite struct {
	ctx  context.Context
	t    *testing.T
	mock Subscription
}

func newE2ETestSuite(ctx context.Context, t *testing.T) *e2eTestSuite {
	return &e2eTestSuite{
		ctx:  ctx,
		t:    t,
		mock: New().NewSubscription("e2e-"),
	}
}

func (s *e2eTestSuite) createNetwork(networkName string) (gcpclient.VoidOperation, error) {
	return s.mock.InsertNetwork(s.ctx, &computepb.InsertNetworkRequest{
		Project: s.mock.ProjectId(),
		NetworkResource: &computepb.Network{
			Name: ptr.To(networkName),
		},
	})
}

func (s *e2eTestSuite) createNetworkOK(networkName string) *computepb.Network {
	opComp, err := s.createNetwork(networkName)
	require.NoError(s.t, err)

	require.NoError(s.t, opComp.Wait(s.ctx))

	net, err := s.mock.GetNetwork(s.ctx, &computepb.GetNetworkRequest{
		Project: s.mock.ProjectId(),
		Network: networkName,
	})
	require.NoError(s.t, err)

	nd, err := gcputil.ParseNameDetail(net.GetSelfLink())
	require.NoError(s.t, err)
	require.Equal(s.t, gcputil.ResourceTypeGlobalNetwork, nd.ResourceType())

	return net
}

func (s *e2eTestSuite) deleteNetwork(networkName string) (gcpclient.VoidOperation, error) {
	return s.mock.DeleteNetwork(s.ctx, &computepb.DeleteNetworkRequest{
		Project: s.mock.ProjectId(),
		Network: networkName,
	})
}

func (s *e2eTestSuite) createPsaRange(networkSelfLink string, addressName string, addressIp string, prefix int32) (gcpclient.VoidOperation, error) {
	return s.mock.InsertGlobalAddress(s.ctx, &computepb.InsertGlobalAddressRequest{
		Project: s.mock.ProjectId(),
		AddressResource: &computepb.Address{
			Name:         ptr.To(addressName),
			Address:      ptr.To(addressIp),
			PrefixLength: ptr.To(prefix),
			Network:      ptr.To(networkSelfLink),
			AddressType:  ptr.To(computepb.Address_INTERNAL.String()),
			Purpose:      ptr.To(computepb.Address_VPC_PEERING.String()),
		},
	})
}

func (s *e2eTestSuite) createPsaRangeOK(networkSelfLink string, addressName string, addressIp string, prefix int32) *computepb.Address {
	opComp, err := s.createPsaRange(networkSelfLink, addressName, addressIp, prefix)
	require.NoError(s.t, err)

	require.NoError(s.t, opComp.Wait(s.ctx))

	addr, err := s.mock.GetGlobalAddress(s.ctx, &computepb.GetGlobalAddressRequest{
		Project: s.mock.ProjectId(),
		Address: addressName,
	})
	require.NoError(s.t, err)

	nd, err := gcputil.ParseNameDetail(addr.GetSelfLink())
	require.NoError(s.t, err)
	require.Equal(s.t, gcputil.ResourceTypeGlobalAddress, nd.ResourceType())

	return addr
}

func (s *e2eTestSuite) deleteAddress(addressName string) (gcpclient.VoidOperation, error) {
	return s.mock.DeleteGlobalAddress(s.ctx, &computepb.DeleteGlobalAddressRequest{
		Project: s.mock.ProjectId(),
		Address: addressName,
	})
}

func (s *e2eTestSuite) deleteAddressOK(addressName string) {
	op, err := s.deleteAddress(addressName)
	require.NoError(s.t, err)
	require.True(s.t, op.Done())
	require.NoError(s.t, op.Wait(s.ctx))
}

func (s *e2eTestSuite) createFilestore(parent string, instanceId string, networkSelfLink string, addressSelfLink string, capacityGb int64) (gcpclient.ResultOperation[*filestorepb.Instance], error) {
	return s.mock.CreateFilestoreInstance(s.ctx, &filestorepb.CreateInstanceRequest{
		Parent:     parent,
		InstanceId: instanceId,
		Instance: &filestorepb.Instance{
			Tier: filestorepb.Instance_PREMIUM,
			FileShares: []*filestorepb.FileShareConfig{
				{
					Name:       "vol1",
					CapacityGb: capacityGb,
				},
			},
			Networks: []*filestorepb.NetworkConfig{
				{
					Network:         networkSelfLink,
					Modes:           []filestorepb.NetworkConfig_AddressMode{filestorepb.NetworkConfig_MODE_IPV4},
					ReservedIpRange: addressSelfLink,
					ConnectMode:     filestorepb.NetworkConfig_PRIVATE_SERVICE_ACCESS,
				},
			},
		},
	})
}

func (s *e2eTestSuite) createFilestoreOK(parent string, instanceId string, networkSelfLink string, addressSelfLink string, capacityGb int64) *filestorepb.Instance {
	opFS, err := s.createFilestore(parent, instanceId, networkSelfLink, addressSelfLink, capacityGb)
	require.NoError(s.t, err)

	require.False(s.t, opFS.Done())

	parentName, err := gcputil.ParseNameDetail(parent)
	require.NoError(s.t, err)

	fsName := gcputil.NewInstanceName(parentName.ProjectId(), parentName.LocationRegionId(), instanceId)

	fs, err := s.mock.GetFilestoreInstance(s.ctx, &filestorepb.GetInstanceRequest{
		Name: fsName.String(),
	})
	require.NoError(s.t, err)
	require.Equal(s.t, filestorepb.Instance_CREATING, fs.State)

	require.NoError(s.t, s.mock.ResolveFilestoreOperation(s.ctx, opFS.Name()))

	fs, err = opFS.Wait(s.ctx)
	require.NoError(s.t, err)
	require.Equal(s.t, filestorepb.Instance_READY, fs.State)

	nd, err := gcputil.ParseNameDetail(fs.Name)
	require.NoError(s.t, err)
	require.Equal(s.t, gcputil.ResourceTypeInstance, nd.ResourceType())

	fs, err = s.mock.GetFilestoreInstance(s.ctx, &filestorepb.GetInstanceRequest{
		Name: nd.String(),
	})
	require.NoError(s.t, err)
	require.Equal(s.t, filestorepb.Instance_READY, fs.State)

	return fs
}

func (s *e2eTestSuite) updateFilestoreOK(update *filestorepb.Instance, updatePaths []string) *filestorepb.Instance {
	opUpdate, err := s.mock.UpdateFilestoreInstance(s.ctx, &filestorepb.UpdateInstanceRequest{
		Instance:   update,
		UpdateMask: &field_mask.FieldMask{Paths: updatePaths},
	})
	require.NoError(s.t, err)
	require.False(s.t, opUpdate.Done())
	require.NoError(s.t, s.mock.ResolveFilestoreOperation(s.ctx, opUpdate.Name()))
	require.True(s.t, opUpdate.Done())

	fs, err := s.mock.GetFilestoreInstance(s.ctx, &filestorepb.GetInstanceRequest{
		Name: update.Name,
	})
	require.NoError(s.t, err)

	return fs
}

func (s *e2eTestSuite) deleteFilestoreOK(instanceName string) {
	op, err := s.mock.DeleteFilestoreInstance(s.ctx, &filestorepb.DeleteInstanceRequest{
		Name: instanceName,
	})
	require.NoError(s.t, err)
	require.False(s.t, op.Done())

	fs, err := s.mock.GetFilestoreInstance(s.ctx, &filestorepb.GetInstanceRequest{
		Name: instanceName,
	})
	require.NoError(s.t, err)
	require.Equal(s.t, filestorepb.Instance_DELETING, fs.State)

	require.NoError(s.t, s.mock.ResolveFilestoreOperation(s.ctx, op.Name()))

	_, err = s.mock.GetFilestoreInstance(s.ctx, &filestorepb.GetInstanceRequest{
		Name: instanceName,
	})
	require.True(s.t, gcpmeta.IsNotFound(err))
}

// redis instance ======================================================================

func (s *e2eTestSuite) createRedisInstance(parent string, instanceId string, networkSelfLink string, addressSelfLink string, memorySizeGb int32) (gcpclient.ResultOperation[*redispb.Instance], error) {
	return s.mock.CreateRedisInstance(s.ctx, &redispb.CreateInstanceRequest{
		Parent:     parent,
		InstanceId: instanceId,
		Instance: &redispb.Instance{
			DisplayName:       instanceId,
			MemorySizeGb:      memorySizeGb,
			Tier:              redispb.Instance_BASIC,
			RedisVersion:      "REDIS_6_X",
			ConnectMode:       redispb.Instance_PRIVATE_SERVICE_ACCESS,
			AuthorizedNetwork: networkSelfLink,
			ReservedIpRange:   addressSelfLink,
		},
	})
}

func (s *e2eTestSuite) createRedisInstanceOK(parent string, instanceId string, networkSelfLink string, addressSelfLink string, memorySizeGb int32) *redispb.Instance {
	op, err := s.createRedisInstance(parent, instanceId, networkSelfLink, addressSelfLink, memorySizeGb)
	require.NoError(s.t, err)
	require.False(s.t, op.Done())

	parentName, err := gcputil.ParseNameDetail(parent)
	require.NoError(s.t, err)

	riName := gcputil.NewInstanceName(parentName.ProjectId(), parentName.LocationRegionId(), instanceId)
	ri, err := s.mock.GetRedisInstance(s.ctx, &redispb.GetInstanceRequest{
		Name: riName.String(),
	})
	require.NoError(s.t, err)
	require.Equal(s.t, redispb.Instance_CREATING, ri.State)

	require.NoError(s.t, s.mock.ResolveRedisInstanceOperation(s.ctx, op.Name()))

	ri, err = op.Wait(s.ctx)
	require.NoError(s.t, err)
	require.Equal(s.t, redispb.Instance_READY, ri.State)

	ri, err = s.mock.GetRedisInstance(s.ctx, &redispb.GetInstanceRequest{
		Name: riName.String(),
	})
	require.NoError(s.t, err)
	require.Equal(s.t, redispb.Instance_READY, ri.State)

	return ri
}

func (s *e2eTestSuite) updateRedisInstanceOK(update *redispb.Instance, updatePaths []string) *redispb.Instance {
	opUpdate, err := s.mock.UpdateRedisInstance(s.ctx, &redispb.UpdateInstanceRequest{
		Instance:   update,
		UpdateMask: &field_mask.FieldMask{Paths: updatePaths},
	})
	require.NoError(s.t, err)
	require.False(s.t, opUpdate.Done())
	require.NoError(s.t, s.mock.ResolveRedisInstanceOperation(s.ctx, opUpdate.Name()))
	require.True(s.t, opUpdate.Done())

	ri, err := s.mock.GetRedisInstance(s.ctx, &redispb.GetInstanceRequest{
		Name: update.Name,
	})
	require.NoError(s.t, err)

	return ri
}

func (s *e2eTestSuite) deleteRedisInstanceOK(instanceName string) {
	op, err := s.mock.DeleteRedisInstance(s.ctx, &redispb.DeleteInstanceRequest{
		Name: instanceName,
	})
	require.NoError(s.t, err)
	require.False(s.t, op.Done())

	ri, err := s.mock.GetRedisInstance(s.ctx, &redispb.GetInstanceRequest{
		Name: instanceName,
	})
	require.NoError(s.t, err)
	require.Equal(s.t, redispb.Instance_DELETING, ri.State)

	require.NoError(s.t, s.mock.ResolveRedisInstanceOperation(s.ctx, op.Name()))

	_, err = s.mock.GetRedisInstance(s.ctx, &redispb.GetInstanceRequest{
		Name: instanceName,
	})
	require.True(s.t, gcpmeta.IsNotFound(err))
}
