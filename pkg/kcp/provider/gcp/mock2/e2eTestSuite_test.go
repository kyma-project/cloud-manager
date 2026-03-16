package mock2

import (
	"context"
	"fmt"
	"testing"

	"cloud.google.com/go/compute/apiv1/computepb"
	"cloud.google.com/go/filestore/apiv1/filestorepb"
	"cloud.google.com/go/networkconnectivity/apiv1/networkconnectivitypb"
	"cloud.google.com/go/redis/apiv1/redispb"
	"cloud.google.com/go/redis/cluster/apiv1/clusterpb"
	"cloud.google.com/go/resourcemanager/apiv3/resourcemanagerpb"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
	gcputil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/util"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/servicenetworking/v1"
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"k8s.io/utils/ptr"
)

type e2eTestSuite struct {
	ctx  context.Context
	t    *testing.T
	mock Store
}

func newE2ETestSuite(ctx context.Context, t *testing.T, mocks ...Server) *e2eTestSuite {
	var mock Server
	if len(mocks) > 0 {
		mock = mocks[0]
	} else {
		mock = New()
	}
	return &e2eTestSuite{
		ctx:  ctx,
		t:    t,
		mock: mock.NewSubscription("e2e-"),
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

func (s *e2eTestSuite) deleteNetworkOK(networkName string) {
	op, err := s.deleteNetwork(networkName)
	require.NoError(s.t, err)
	require.True(s.t, op.Done())
	require.NoError(s.t, op.Wait(s.ctx))

	_, err = s.mock.GetNetwork(s.ctx, &computepb.GetNetworkRequest{
		Project: s.mock.ProjectId(),
		Network: networkName,
	})
	require.Error(s.t, err)
}

// Peering ==================================================================================

func (s *e2eTestSuite) addPeeringOK(localNetwork, peeringName, remoteNetwork string) *computepb.NetworkPeering {
	netNd, err := gcputil.ParseNameDetail(localNetwork)
	require.NoError(s.t, err)

	op, err := s.mock.AddPeering(s.ctx, &computepb.AddPeeringNetworkRequest{
		Network: netNd.ResourceId(),
		Project: s.mock.ProjectId(),
		NetworksAddPeeringRequestResource: &computepb.NetworksAddPeeringRequest{
			NetworkPeering: &computepb.NetworkPeering{
				Name:    ptr.To(peeringName),
				Network: ptr.To(remoteNetwork),
			},
		},
	})
	require.NoError(s.t, err)
	require.True(s.t, op.Done())
	require.NoError(s.t, op.Wait(s.ctx))

	peering := s.getPeering(netNd.String(), peeringName)
	require.NotNil(s.t, peering)

	return peering
}

func (s *e2eTestSuite) removePeeringOK(localNetwork, peeringName string) {
	netNd, err := gcputil.ParseNameDetail(localNetwork)
	require.NoError(s.t, err)

	op, err := s.mock.RemovePeering(s.ctx, &computepb.RemovePeeringNetworkRequest{
		Project: s.mock.ProjectId(),
		Network: netNd.ResourceId(),
		NetworksRemovePeeringRequestResource: &computepb.NetworksRemovePeeringRequest{
			Name: ptr.To(peeringName),
		},
	})
	require.NoError(s.t, err)
	require.True(s.t, op.Done())
	require.NoError(s.t, op.Wait(s.ctx))

	net, err := s.mock.GetNetwork(s.ctx, &computepb.GetNetworkRequest{
		Project: s.mock.ProjectId(),
		Network: netNd.ResourceId(),
	})
	require.NoError(s.t, err)
	require.NotNil(s.t, net)

	peeringFound := false
	for _, p := range net.GetPeerings() {
		if p.GetName() == peeringName {
			peeringFound = true
			break
		}
	}
	require.False(s.t, peeringFound)
}

func (s *e2eTestSuite) getPeering(network, peering string) *computepb.NetworkPeering {
	netNd, err := gcputil.ParseNameDetail(network)
	require.NoError(s.t, err)

	net, err := s.mock.GetNetwork(s.ctx, &computepb.GetNetworkRequest{
		Project: s.mock.ProjectId(),
		Network: netNd.ResourceId(),
	})
	require.NoError(s.t, err)
	require.NotNil(s.t, net)

	for _, p := range net.GetPeerings() {
		if p.GetName() == peering {
			return p
		}
	}
	return nil
}

// Subnet ==================================================================================

func (s *e2eTestSuite) createSubnet(region, networkName, subnetId, ipCidrRange string) (gcpclient.VoidOperation, error) {
	return s.mock.InsertSubnet(s.ctx, &computepb.InsertSubnetworkRequest{
		Project: s.mock.ProjectId(),
		Region:  region,
		SubnetworkResource: &computepb.Subnetwork{
			Name:        ptr.To(subnetId),
			IpCidrRange: ptr.To(ipCidrRange),
			Network:     ptr.To(networkName),
			Region:      ptr.To(region),
		},
	})
}

func (s *e2eTestSuite) createSubnetOK(region, networkName, subnetId, ipCidrRange string) *computepb.Subnetwork {
	op, err := s.createSubnet(region, networkName, subnetId, ipCidrRange)
	require.NoError(s.t, err)
	require.True(s.t, op.Done())

	sub, err := s.mock.GetSubnet(s.ctx, &computepb.GetSubnetworkRequest{
		Project:    s.mock.ProjectId(),
		Region:     region,
		Subnetwork: subnetId,
	})
	require.NoError(s.t, err)
	require.Equal(s.t, computepb.Subnetwork_READY.String(), sub.GetState())

	ndSubnet, err := gcputil.ParseNameDetail(sub.GetSelfLink())
	require.NoError(s.t, err)
	require.Equal(s.t, s.mock.ProjectId(), ndSubnet.ProjectId())
	require.Equal(s.t, region, ndSubnet.LocationRegionId())
	require.Equal(s.t, subnetId, ndSubnet.ResourceId())
	require.Equal(s.t, gcputil.ResourceTypeSubnetwork, ndSubnet.ResourceType())

	return sub
}

func (s *e2eTestSuite) deleteSubnet(region, subnetId string) (gcpclient.VoidOperation, error) {
	return s.mock.DeleteSubnet(s.ctx, &computepb.DeleteSubnetworkRequest{
		Project:    s.mock.ProjectId(),
		Region:     region,
		Subnetwork: subnetId,
	})
}

func (s *e2eTestSuite) deleteSubnetOK(region, subnetId string) {
	op, err := s.deleteSubnet(region, subnetId)
	require.NoError(s.t, err)
	require.True(s.t, op.Done())

	_, err = s.mock.GetSubnet(s.ctx, &computepb.GetSubnetworkRequest{
		Project:    s.mock.ProjectId(),
		Region:     region,
		Subnetwork: subnetId,
	})
	require.Error(s.t, err)
}

// Router ================================================================================

func (s *e2eTestSuite) createRouter(region, networkName, routerId string, subnetNames []string) (gcpclient.VoidOperation, error) {
	req := &computepb.InsertRouterRequest{
		Project: s.mock.ProjectId(),
		Region:  region,
		RouterResource: &computepb.Router{
			Name:    ptr.To(routerId),
			Network: ptr.To(networkName),
			Region:  ptr.To(region),
			Nats: []*computepb.RouterNat{
				{
					Name:                          ptr.To(routerId),
					NatIpAllocateOption:           ptr.To(computepb.RouterNat_AUTO_ONLY.String()),
					SourceSubnetworkIpRangesToNat: ptr.To(computepb.RouterNat_LIST_OF_SUBNETWORKS.String()),
				},
			},
		},
	}
	for _, subNameTxt := range subnetNames {
		ss := &computepb.RouterNatSubnetworkToNat{
			Name:                ptr.To(subNameTxt),
			SourceIpRangesToNat: []string{computepb.RouterNat_ALL_SUBNETWORKS_ALL_IP_RANGES.String()},
		}
		req.RouterResource.Nats[0].Subnetworks = append(req.RouterResource.Nats[0].Subnetworks, ss)
	}
	return s.mock.InsertRouter(s.ctx, req)
}

func (s *e2eTestSuite) createRouterOK(region, networkName, routerId string, subnetNames []string) *computepb.Router {
	op, err := s.createRouter(region, networkName, routerId, subnetNames)
	require.NoError(s.t, err)
	require.True(s.t, op.Done())

	rt, err := s.mock.GetRouter(s.ctx, &computepb.GetRouterRequest{
		Project: s.mock.ProjectId(),
		Region:  region,
		Router:  routerId,
	})
	require.NoError(s.t, err)

	return rt
}

func (s *e2eTestSuite) deleteRouter(region, routerId string) (gcpclient.VoidOperation, error) {
	return s.mock.DeleteRouter(s.ctx, &computepb.DeleteRouterRequest{
		Project: s.mock.ProjectId(),
		Region:  region,
		Router:  routerId,
	})
}

func (s *e2eTestSuite) deleteRouterOK(region, routerId string) {
	op, err := s.deleteRouter(region, routerId)
	require.NoError(s.t, err)
	require.True(s.t, op.Done())

	_, err = s.mock.GetRouter(s.ctx, &computepb.GetRouterRequest{
		Project: s.mock.ProjectId(),
		Region:  region,
		Router:  routerId,
	})
	require.Error(s.t, err)
}

// PSA IP Range ============================================================================

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

// PSA connection ==========================================================================

func (s *e2eTestSuite) createPsaConnectionOK(networkLink, addrLink string) *servicenetworking.Connection {
	netNd, err := gcputil.ParseNameDetail(networkLink)
	require.NoError(s.t, err)

	addrNd, err := gcputil.ParseNameDetail(addrLink)
	require.NoError(s.t, err)

	op, err := s.mock.CreateServiceConnection(s.ctx, netNd.ProjectId(), netNd.ResourceId(), []string{addrNd.ResourceId()})

	require.NoError(s.t, err)
	require.True(s.t, op.Done)

	arr, err := s.mock.ListServiceConnections(s.ctx, netNd.ProjectId(), netNd.ResourceId())
	require.NoError(s.t, err)
	require.Len(s.t, arr, 1)

	return arr[0]
}

func (s *e2eTestSuite) deletePsaConnectionOK(networkLink string) {
	netNd, err := gcputil.ParseNameDetail(networkLink)
	require.NoError(s.t, err)

	op, err := s.mock.DeleteServiceConnection(s.ctx, netNd.ProjectId(), netNd.ResourceId())
	require.NoError(s.t, err)
	require.True(s.t, op.Done)

	arr, err := s.mock.ListServiceConnections(s.ctx, s.mock.ProjectId(), netNd.ResourceId())
	require.NoError(s.t, err)
	require.Len(s.t, arr, 0)
}

// Filestore =====================================================================================

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

// filestore backup ====================================================================

func (s *e2eTestSuite) createFilestoreBackup(parent, backupId, sourceInstance, sourceFileShare string, labels map[string]string) (gcpclient.ResultOperation[*filestorepb.Backup], error) {
	return s.mock.CreateFilestoreBackup(s.ctx, &filestorepb.CreateBackupRequest{
		Parent:   parent,
		BackupId: backupId,
		Backup: &filestorepb.Backup{
			Labels:          labels,
			SourceInstance:  sourceInstance,
			SourceFileShare: sourceFileShare,
		},
	})
}

func (s *e2eTestSuite) createFilestoreBackupOK(parent, backupId, sourceInstance, sourceFileShare string, labels map[string]string) *filestorepb.Backup {
	op, err := s.createFilestoreBackup(parent, backupId, sourceInstance, sourceFileShare, labels)
	require.NoError(s.t, err)
	require.False(s.t, op.Done())

	parentNd, err := gcputil.ParseNameDetail(parent)
	require.NoError(s.t, err)

	backupNd := gcputil.NewBackupName(parentNd.ProjectId(), parentNd.LocationRegionId(), backupId)

	backup, err := s.mock.GetFilestoreBackup(s.ctx, &filestorepb.GetBackupRequest{
		Name: backupNd.String(),
	})
	require.NoError(s.t, err)
	require.Equal(s.t, filestorepb.Backup_CREATING, backup.State)

	err = s.mock.ResolveFilestoreBackupOperation(s.ctx, op.Name())
	require.NoError(s.t, err)

	backup, err = op.Wait(s.ctx)
	require.NoError(s.t, err)
	require.Equal(s.t, filestorepb.Backup_READY, backup.State)

	backup, err = s.mock.GetFilestoreBackup(s.ctx, &filestorepb.GetBackupRequest{
		Name: backupNd.String(),
	})
	require.NoError(s.t, err)
	require.Equal(s.t, filestorepb.Backup_READY, backup.State)
	require.Equal(s.t, labels, backup.Labels)

	require.Equal(s.t, backupNd.String(), backup.Name)

	return backup
}

func (s *e2eTestSuite) updateFilestoreBackupLabels(backupName string, labels map[string]string) (gcpclient.ResultOperation[*filestorepb.Backup], error) {
	return s.mock.UpdateFilestoreBackup(s.ctx, &filestorepb.UpdateBackupRequest{
		Backup: &filestorepb.Backup{
			Name:   backupName,
			Labels: labels,
		},
		UpdateMask: &field_mask.FieldMask{Paths: []string{"labels"}},
	})
}

func (s *e2eTestSuite) updateFilestoreBackupLabelsOK(backupName string, labels map[string]string) *filestorepb.Backup {
	op, err := s.updateFilestoreBackupLabels(backupName, labels)
	require.NoError(s.t, err)
	require.True(s.t, op.Done())

	backup, err := s.mock.GetFilestoreBackup(s.ctx, &filestorepb.GetBackupRequest{
		Name: backupName,
	})
	require.NoError(s.t, err)
	require.Equal(s.t, labels, backup.Labels)

	return backup
}

func (s *e2eTestSuite) deleteFilestoreBackup(backupName string) (gcpclient.VoidOperation, error) {
	return s.mock.DeleteFilestoreBackup(s.ctx, &filestorepb.DeleteBackupRequest{
		Name: backupName,
	})
}

func (s *e2eTestSuite) deleteFilestoreBackupOK(backupName string) {
	op, err := s.deleteFilestoreBackup(backupName)
	require.NoError(s.t, err)
	require.False(s.t, op.Done())

	backup, err := s.mock.GetFilestoreBackup(s.ctx, &filestorepb.GetBackupRequest{
		Name: backupName,
	})
	require.NoError(s.t, err)
	require.Equal(s.t, filestorepb.Backup_DELETING, backup.State)

	err = s.mock.ResolveFilestoreBackupOperation(s.ctx, op.Name())
	require.NoError(s.t, err)

	_, err = s.mock.GetFilestoreBackup(s.ctx, &filestorepb.GetBackupRequest{
		Name: backupName,
	})
	require.Error(s.t, err)
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

// ServiceConnectionPolicy ==========================================================

func (s *e2eTestSuite) createServiceConnectionPolicy(parent, serviceConnectionPolicyId, network string, subnetworks []string) (gcpclient.ResultOperation[*networkconnectivitypb.ServiceConnectionPolicy], error) {
	return s.mock.CreateServiceConnectionPolicy(s.ctx, &networkconnectivitypb.CreateServiceConnectionPolicyRequest{
		Parent:                    parent,
		ServiceConnectionPolicyId: serviceConnectionPolicyId,
		ServiceConnectionPolicy: &networkconnectivitypb.ServiceConnectionPolicy{
			Network:      network,
			ServiceClass: "gcp-memorystore-redis",
			PscConfig: &networkconnectivitypb.ServiceConnectionPolicy_PscConfig{
				Subnetworks: subnetworks,
			},
		},
	})
}

func (s *e2eTestSuite) createServiceConnectionPolicyOK(parent, serviceConnectionPolicyId, network string, subnetworks []string) *networkconnectivitypb.ServiceConnectionPolicy {
	parentName, err := gcputil.ParseNameDetail(parent)
	require.NoError(s.t, err)
	require.Equal(s.t, gcputil.ResourceTypeLocation, parentName.ResourceType())

	scpName := gcputil.NewServiceConnectionPolicyName(parentName.ProjectId(), parentName.LocationRegionId(), serviceConnectionPolicyId)

	op, err := s.createServiceConnectionPolicy(parent, serviceConnectionPolicyId, network, subnetworks)
	require.NoError(s.t, err)
	require.True(s.t, op.Done())

	scp, err := op.Wait(s.ctx)
	require.NoError(s.t, err)
	scpNetworkName, err := gcputil.ParseNameDetail(scp.Network)
	require.NoError(s.t, err)
	require.Equal(s.t, gcputil.ResourceTypeGlobalNetwork, scpNetworkName.ResourceType())
	require.True(s.t, scpNetworkName.EqualString(network))

	require.Equal(s.t, len(subnetworks), len(scp.PscConfig.Subnetworks))
	for _, subnetTxt := range scp.PscConfig.Subnetworks {
		subnetName, err := gcputil.ParseNameDetail(subnetTxt)
		require.NoError(s.t, err)
		require.Equal(s.t, gcputil.ResourceTypeSubnetwork, subnetName.ResourceType())
		match := false
		for _, s := range subnetworks {
			if subnetName.EqualString(s) {
				match = true
				break
			}
		}
		require.True(s.t, match, "subnet %s not found", subnetTxt)
	}
	require.Equal(s.t, subnetworks, scp.PscConfig.Subnetworks)

	scp, err = s.mock.GetServiceConnectionPolicy(s.ctx, &networkconnectivitypb.GetServiceConnectionPolicyRequest{
		Name: scpName.String(),
	})
	require.NoError(s.t, err)
	scpNetworkName, err = gcputil.ParseNameDetail(scp.Network)
	require.NoError(s.t, err)
	require.Equal(s.t, gcputil.ResourceTypeGlobalNetwork, scpNetworkName.ResourceType())
	require.True(s.t, scpNetworkName.EqualString(network))

	require.Equal(s.t, len(subnetworks), len(scp.PscConfig.Subnetworks))

	return scp
}

func (s *e2eTestSuite) updateServiceConnectionPolicy(scp *networkconnectivitypb.ServiceConnectionPolicy, updateMask []string) (gcpclient.ResultOperation[*networkconnectivitypb.ServiceConnectionPolicy], error) {
	return s.mock.UpdateServiceConnectionPolicy(s.ctx, &networkconnectivitypb.UpdateServiceConnectionPolicyRequest{
		ServiceConnectionPolicy: scp,
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: updateMask,
		},
	})
}

func (s *e2eTestSuite) updateServiceConnectionPolicyOK(scp *networkconnectivitypb.ServiceConnectionPolicy, updateMask []string) *networkconnectivitypb.ServiceConnectionPolicy {
	op, err := s.updateServiceConnectionPolicy(scp, updateMask)
	require.NoError(s.t, err)
	require.True(s.t, op.Done())

	scp, err = op.Wait(s.ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, scp)

	scp, err = s.mock.GetServiceConnectionPolicy(s.ctx, &networkconnectivitypb.GetServiceConnectionPolicyRequest{
		Name: scp.Name,
	})
	require.NoError(s.t, err)
	require.NotNil(s.t, scp)
	return scp
}

func (s *e2eTestSuite) deleteServiceConnectionPolicy(serviceConnectionPolicyName string) (gcpclient.VoidOperation, error) {
	return s.mock.DeleteServiceConnectionPolicy(s.ctx, &networkconnectivitypb.DeleteServiceConnectionPolicyRequest{
		Name: serviceConnectionPolicyName,
	})
}

func (s *e2eTestSuite) deleteServiceConnectionPolicyOK(serviceConnectionPolicyName string) {
	op, err := s.deleteServiceConnectionPolicy(serviceConnectionPolicyName)
	require.NoError(s.t, err)
	require.True(s.t, op.Done())

	_, err = s.mock.GetServiceConnectionPolicy(s.ctx, &networkconnectivitypb.GetServiceConnectionPolicyRequest{
		Name: serviceConnectionPolicyName,
	})
	require.Error(s.t, err)
	require.True(s.t, gcpmeta.IsNotFound(err))
}

// redis cluster ======================================================================

func (s *e2eTestSuite) createRedisCluster(parent, network, clusterId string, replicaCount, shardCount int32, redisConfigs map[string]string) (gcpclient.ResultOperation[*clusterpb.Cluster], error) {
	return s.mock.CreateRedisCluster(s.ctx, &clusterpb.CreateClusterRequest{
		Parent:    parent,
		ClusterId: clusterId,
		Cluster: &clusterpb.Cluster{
			ReplicaCount: ptr.To(replicaCount),
			ShardCount:   ptr.To(shardCount),
			NodeType:     clusterpb.NodeType_REDIS_STANDARD_SMALL,
			PscConfigs: []*clusterpb.PscConfig{{
				Network: network,
			}},
			RedisConfigs:              redisConfigs,
			PersistenceConfig:         &clusterpb.ClusterPersistenceConfig{Mode: clusterpb.ClusterPersistenceConfig_DISABLED},
			AuthorizationMode:         clusterpb.AuthorizationMode_AUTH_MODE_DISABLED,
			TransitEncryptionMode:     clusterpb.TransitEncryptionMode_TRANSIT_ENCRYPTION_MODE_SERVER_AUTHENTICATION,
			ZoneDistributionConfig:    &clusterpb.ZoneDistributionConfig{Mode: clusterpb.ZoneDistributionConfig_MULTI_ZONE},
			DeletionProtectionEnabled: ptr.To(false),
		},
	})
}

func (s *e2eTestSuite) createRedisClusterOK(parent, network, clusterId string, replicaCount, shardCount int32, redisConfigs map[string]string) *clusterpb.Cluster {
	op, err := s.createRedisCluster(parent, network, clusterId, replicaCount, shardCount, redisConfigs)
	require.NoError(s.t, err)
	require.False(s.t, op.Done())

	parentName, err := gcputil.ParseNameDetail(parent)
	require.NoError(s.t, err)

	rcName := gcputil.NewClusterName(parentName.ProjectId(), parentName.LocationRegionId(), clusterId)
	rc, err := s.mock.GetRedisCluster(s.ctx, &clusterpb.GetClusterRequest{
		Name: rcName.String(),
	})
	require.NoError(s.t, err)
	require.Equal(s.t, clusterpb.Cluster_CREATING, rc.State)

	err = s.mock.ResolveRedisClusterOperation(s.ctx, op.Name())
	require.NoError(s.t, err)

	rc, err = op.Wait(s.ctx)
	require.NoError(s.t, err)
	require.Equal(s.t, clusterpb.Cluster_ACTIVE, rc.State)

	rc, err = s.mock.GetRedisCluster(s.ctx, &clusterpb.GetClusterRequest{
		Name: rcName.String(),
	})
	require.NoError(s.t, err)
	require.Equal(s.t, clusterpb.Cluster_ACTIVE, rc.State)

	return rc
}

func (s *e2eTestSuite) deleteRedisClusterOK(clusterName string) {
	op, err := s.mock.DeleteRedisCluster(s.ctx, &clusterpb.DeleteClusterRequest{
		Name: clusterName,
	})
	require.NoError(s.t, err)
	require.False(s.t, op.Done())

	rc, err := s.mock.GetRedisCluster(s.ctx, &clusterpb.GetClusterRequest{
		Name: clusterName,
	})
	require.NoError(s.t, err)
	require.Equal(s.t, clusterpb.Cluster_DELETING, rc.State)

	err = s.mock.ResolveRedisClusterOperation(s.ctx, op.Name())
	require.NoError(s.t, err)

	_, err = s.mock.GetRedisCluster(s.ctx, &clusterpb.GetClusterRequest{
		Name: clusterName,
	})
	require.Error(s.t, err)
	require.True(s.t, gcpmeta.IsNotFound(err))
}

// resource manager tags ==========================================================

func (s *e2eTestSuite) createTagKey(keyShortName string) *resourcemanagerpb.TagKey {
	parentNd := gcputil.NewProjectName(s.mock.ProjectId())
	op, err := s.mock.CreateTagKey(s.ctx, &resourcemanagerpb.CreateTagKeyRequest{
		TagKey: &resourcemanagerpb.TagKey{
			Parent:    parentNd.String(),
			ShortName: keyShortName,
		},
	})
	require.NoError(s.t, err)
	require.True(s.t, op.Done())
	key, err := op.Wait(s.ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, key)
	require.Equal(s.t, keyShortName, key.ShortName)

	k := s.mock.GetTagKeyByShortNameNoLock(keyShortName)
	require.NotNil(s.t, k)
	require.Equal(s.t, k.ShortName, key.ShortName)
	require.Equal(s.t, k.Parent, key.Parent)
	require.Equal(s.t, k.Name, key.Name)
	require.Equal(s.t, fmt.Sprintf("%s/%s", s.mock.ProjectId(), keyShortName), key.NamespacedName)

	return key
}

func (s *e2eTestSuite) deleteTagKey(keyFullName string) *resourcemanagerpb.TagKey {
	op, err := s.mock.DeleteTagKey(s.ctx, &resourcemanagerpb.DeleteTagKeyRequest{
		Name: keyFullName,
	})
	require.NoError(s.t, err)
	require.True(s.t, op.Done())
	key, err := op.Wait(s.ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, key)
	require.Equal(s.t, keyFullName, key.Name)
	require.Equal(s.t, fmt.Sprintf("%s/%s", s.mock.ProjectId(), key.ShortName), key.NamespacedName)

	k := s.mock.GetTagKeyByShortNameNoLock(key.ShortName)
	require.Nil(s.t, k)

	return key
}

func (s *e2eTestSuite) createTagValue(keyShortName, valueShortName string) *resourcemanagerpb.TagValue {
	key := s.mock.GetTagKeyByShortNameNoLock(keyShortName)
	require.NotNil(s.t, key)

	op, err := s.mock.CreateTagValue(s.ctx, &resourcemanagerpb.CreateTagValueRequest{
		TagValue: &resourcemanagerpb.TagValue{
			Parent:    key.Name,
			ShortName: valueShortName,
		},
	})
	require.NoError(s.t, err)
	require.True(s.t, op.Done())
	value, err := op.Wait(s.ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, value)
	require.Equal(s.t, key.Name, value.Parent)
	require.Equal(s.t, valueShortName, value.ShortName)
	require.Equal(s.t, fmt.Sprintf("%s/%s/%s", s.mock.ProjectId(), key.ShortName, value.ShortName), value.NamespacedName)

	v := s.mock.GetTagValueByShortNameNoLock(key.Name, value.ShortName)
	require.NotNil(s.t, v)
	require.Equal(s.t, value.Name, v.Name)

	return value
}

func (s *e2eTestSuite) deleteTagValue(valueFullName string) {
	op, err := s.mock.DeleteTagValue(s.ctx, &resourcemanagerpb.DeleteTagValueRequest{
		Name: valueFullName,
	})
	require.NoError(s.t, err)
	require.True(s.t, op.Done())
	value, err := op.Wait(s.ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, value)
	require.Equal(s.t, valueFullName, value.Name)

	v := s.mock.GetTagValueByFullNameNoLock(value.Name)
	require.Nil(s.t, v)
}

func (s *e2eTestSuite) createTagBinding(target, valueFullName string) *resourcemanagerpb.TagBinding {
	value := s.mock.GetTagValueByFullNameNoLock(valueFullName)
	require.NotNilf(s.t, value, "tag value %s does not exist", valueFullName)
	key, err := s.mock.GetTagKey(s.ctx, &resourcemanagerpb.GetTagKeyRequest{
		Name: value.Parent,
	})
	require.NoError(s.t, err)

	op, err := s.mock.CreateTagBinding(s.ctx, &resourcemanagerpb.CreateTagBindingRequest{
		TagBinding: &resourcemanagerpb.TagBinding{
			Parent:   target,
			TagValue: valueFullName,
		},
	})
	require.NoError(s.t, err)
	require.True(s.t, op.Done())
	binding, err := op.Wait(s.ctx)
	require.NoError(s.t, err)
	require.NotNil(s.t, binding)
	require.Equal(s.t, target, binding.Parent)
	require.Equal(s.t, valueFullName, binding.TagValue)
	require.Equal(s.t, fmt.Sprintf("%s/%s/%s", s.mock.ProjectId(), key.ShortName, value.ShortName), binding.TagValueNamespacedName)

	b := s.mock.GetTagBindingNoLock(binding.Name)
	require.NotNil(s.t, b)

	return binding
}

func (s *e2eTestSuite) deleteTagBinding(bindingName string) {
	binding := s.mock.GetTagBindingNoLock(bindingName)
	require.NotNil(s.t, binding)

	op, err := s.mock.DeleteTagBinding(s.ctx, &resourcemanagerpb.DeleteTagBindingRequest{
		Name: bindingName,
	})
	require.NoError(s.t, err)
	require.True(s.t, op.Done())

	binding = s.mock.GetTagBindingNoLock(binding.Name)
	require.Nil(s.t, binding)
}
