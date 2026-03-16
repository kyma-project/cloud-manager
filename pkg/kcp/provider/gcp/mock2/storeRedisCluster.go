package mock2

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/compute/apiv1/computepb"
	"cloud.google.com/go/redis/cluster/apiv1/clusterpb"
	"github.com/google/uuid"
	"github.com/googleapis/gax-go/v2"
	"github.com/kyma-project/cloud-manager/pkg/common"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
	gcputil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"google.golang.org/protobuf/types/known/timestamppb"
	"k8s.io/utils/ptr"
)

/*
authorizationMode: AUTH_MODE_DISABLED
automatedBackupConfig:
  automatedBackupMode: DISABLED
clusterEndpoints:
- connections:
  - pscAutoConnection:
      address: 10.128.0.36
      connectionType: CONNECTION_TYPE_DISCOVERY
      forwardingRule: https://www.googleapis.com/compute/v1/projects/my-project/regions/us-central1/forwardingRules/sca-auto-fr-ad12345a-1234-1234-1234-12347cd81234
      network: projects/my-project/global/networks/my-network
      projectId: my-project
      pscConnectionId: '28856785926357028'
      pscConnectionStatus: PSC_CONNECTION_STATUS_ACTIVE
      serviceAttachment: projects/818549193598/regions/us-central1/serviceAttachments/gcp-memorystore-auto-oc12341234123412-psc-sa
  - pscAutoConnection:
      address: 10.128.0.11
      forwardingRule: https://www.googleapis.com/compute/v1/projects/my-project/regions/us-central1/forwardingRules/sca-auto-fr-12341234-1234-1234-1234-123412341234
      network: projects/my-project/global/networks/my-network
      projectId: my-project
      pscConnectionId: '12341234123412341'
      pscConnectionStatus: PSC_CONNECTION_STATUS_ACTIVE
      serviceAttachment: projects/818549193598/regions/us-central1/serviceAttachments/gcp-memorystore-auto-oc12341234123412-psc-sa-2
createTime: '2026-03-09T12:18:43.058117330Z'
deletionProtectionEnabled: false
discoveryEndpoints:
- address: 10.128.0.11
  port: 6379
  pscConfig:
    network: projects/my-project/global/networks/my-network
effectiveMaintenanceVersion: REDISCLUSTER_20251008_00_00
encryptionInfo:
  encryptionType: GOOGLE_DEFAULT_ENCRYPTION
name: projects/my-project/locations/us-central1/clusters/my-cluster
nodeType: REDIS_STANDARD_SMALL
persistenceConfig:
  mode: DISABLED
preciseSizeGb: 65.0
pscConnections:
- address: 10.128.0.11
  forwardingRule: https://www.googleapis.com/compute/v1/projects/my-project/regions/us-central1/forwardingRules/sca-auto-fr-12341234-1234-1234-1234-123412341234
  network: projects/my-project/global/networks/my-network
  projectId: my-project
  pscConnectionId: '12341234123457028'
  serviceAttachment: projects/818549193598/regions/us-central1/serviceAttachments/gcp-memorystore-auto-oc12341234123412-psc-sa
- address: 10.128.0.54
  forwardingRule: https://www.googleapis.com/compute/v1/projects/my-project/regions/us-central1/forwardingRules/sca-auto-fr-12341234-1234-abcd-1234-1234abcd1234
  network: projects/my-project/global/networks/my-network
  projectId: my-project
  pscConnectionId: '1234123412347046'
  serviceAttachment: projects/818549193598/regions/us-central1/serviceAttachments/gcp-memorystore-auto-oc12341234123412-psc-sa-2
pscServiceAttachments:
- connectionType: CONNECTION_TYPE_DISCOVERY
  serviceAttachment: projects/818549193598/regions/us-central1/serviceAttachments/gcp-memorystore-auto-oc12341234123412-psc-sa
- serviceAttachment: projects/818549193598/regions/us-central1/serviceAttachments/gcp-memorystore-auto-oc12341234123412-psc-sa-2
replicaCount: 1
satisfiesPzi: false
serverCaMode: SERVER_CA_MODE_UNSPECIFIED
shardCount: 10
sizeGb: 65
state: ACTIVE
transitEncryptionMode: TRANSIT_ENCRYPTION_MODE_DISABLED
uid: 12341234-1234-abcd-1234-1234eea31234
zoneDistributionConfig:
  mode: MULTI_ZONE
*/

// private methods ============================================================================

func (s *store) getRedisClusterNoLock(name string) (*clusterpb.Cluster, error) {
	for _, item := range s.redisClusters.items {
		if item.Name.EqualString(name) {
			return item.Obj, nil
		}
	}
	return nil, gcpmeta.NewNotFoundError("redisCluster %s not found", name)
}

// RedisCluster public client methods ========================================================

func (s *store) CreateRedisCluster(ctx context.Context, req *clusterpb.CreateClusterRequest, _ ...gax.CallOption) (gcpclient.ResultOperation[*clusterpb.Cluster], error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	parentName, err := gcputil.ParseNameDetail(req.Parent)
	if err != nil {
		return nil, gcpmeta.NewBadRequestError("invalid parent: %v", err)
	}
	if parentName.ResourceType() != gcputil.ResourceTypeLocation {
		return nil, gcpmeta.NewBadRequestError("invalid resource type, expected location got %q", parentName.ResourceType())
	}

	// name validation
	if req.ClusterId == "" {
		return nil, gcpmeta.NewBadRequestError("clusterId is required")
	}
	rcName := gcputil.NewClusterName(parentName.ProjectId(), parentName.LocationRegionId(), req.ClusterId)
	_, err = s.getRedisClusterNoLock(rcName.String())
	if err == nil {
		return nil, gcpmeta.NewBadRequestError("redisCluster %s already exists", rcName.String())
	}

	if sc := ptr.Deref(req.Cluster.ShardCount, 1); sc < 1 || sc > 100 {
		return nil, gcpmeta.NewBadRequestError("invalid cluster shard count: %d", ptr.Deref(req.Cluster.ShardCount, 1))
	}
	if rc := ptr.Deref(req.Cluster.ReplicaCount, 1); rc < 1 || rc > 100 {
		return nil, gcpmeta.NewBadRequestError("invalid cluster replica count: %d", ptr.Deref(req.Cluster.ReplicaCount, 1))
	}

	// network => ip
	endpoints := map[string]string{}

	// psc & network validation
	for _, pscConfig := range req.Cluster.PscConfigs {
		netName, err := gcputil.ParseNameDetail(pscConfig.Network)
		if err != nil {
			return nil, gcpmeta.NewBadRequestError("invalid pscConfig network %s: %v", pscConfig.Network, err)
		}
		if netName.ResourceType() != gcputil.ResourceTypeGlobalNetwork {
			return nil, gcpmeta.NewBadRequestError("invalid pscConfig network %s, expected network type, got %s", pscConfig.Network, netName.ResourceType())
		}

		subnetList := s.subnets.FilterByCallback(func(item FilterableListItem[*computepb.Subnetwork]) bool {
			if item.Name.ProjectId() != parentName.ProjectId() {
				return false
			}
			if item.Name.LocationRegionId() != parentName.LocationRegionId() {
				return false
			}
			if netName.EqualString(ptr.Deref(item.Obj.Network, "")) {
				return true
			}
			return false
		})

		scpFound := false
		for _, subnetItem := range subnetList.items {
			scpList, err := s.serviceConnectionPolicies.FilterByExpression(ptr.To(fmt.Sprintf(
				`(psc_config.subnetworks:("%s"))(service_class="gcp-memorystore-redis")`,
				subnetItem.Name.String(),
			)))
			if err != nil {
				return nil, gcpmeta.NewBadRequestError("%v: failed filtering scp using subnetwork: %v", common.ErrLogical, err)
			}
			if scpList.Len() > 0 {
				// simplified, taking first ip of the subnet cidr range
				parts := strings.Split(subnetItem.Obj.GetIpCidrRange(), "/")
				endpoints[netName.String()] = parts[0]
				scpFound = true
				break
			}
		}
		if !scpFound {
			return nil, gcpmeta.NewBadRequestError("network %s does not have SCP in region %s for service class 'gcp-memorystore-redis'", netName.String(), parentName.LocationRegionId())
		}
	}

	rc, err := util.Clone(req.Cluster)
	if err != nil {
		return nil, gcpmeta.NewInternalServerError("%v: failed to clone redisCluster: %v", common.ErrLogical, err)
	}
	rc.Uid = uuid.NewString()
	rc.Name = rcName.String()
	rc.CreateTime = timestamppb.Now()
	rc.State = clusterpb.Cluster_CREATING
	rc.SizeGb = ptr.To(int32(100))
	for netNameTxt, ip := range endpoints {
		rc.DiscoveryEndpoints = append(rc.DiscoveryEndpoints, &clusterpb.DiscoveryEndpoint{
			Address: ip,
			Port:    6378,
			PscConfig: &clusterpb.PscConfig{
				Network: netNameTxt,
			},
		})
	}

	s.redisClusters.Add(rc, rcName)

	opName := s.newLongRunningOperationName()
	b := NewOperationLongRunningBuilder(opName.String(), rcName)
	if err := b.WithRedisClusterMetadata(rcName, "create"); err != nil {
		return nil, fmt.Errorf("%w: failed setting create redisCluster operation metadata: %w", common.ErrLogical, err)
	}
	s.longRunningOperations.Add(b, opName)

	return NewResultOperation[*clusterpb.Cluster](b.GetOperationPB()), nil
}

func (s *store) GetRedisCluster(ctx context.Context, req *clusterpb.GetClusterRequest, _ ...gax.CallOption) (*clusterpb.Cluster, error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	rc, err := s.getRedisClusterNoLock(req.Name)
	if err != nil {
		return nil, err
	}
	return util.Clone(rc)
}

func (s *store) GetRedisClusterCertificateAuthority(ctx context.Context, req *clusterpb.GetClusterCertificateAuthorityRequest, _ ...gax.CallOption) (*clusterpb.CertificateAuthority, error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	return &clusterpb.CertificateAuthority{
		ServerCa: &clusterpb.CertificateAuthority_ManagedServerCa{
			ManagedServerCa: &clusterpb.CertificateAuthority_ManagedCertificateAuthority{
				CaCerts: []*clusterpb.CertificateAuthority_ManagedCertificateAuthority_CertChain{
					{
						Certificates: []string{"-----BEGIN CERTIFICATE-----\nc29tZSBjZXJ0aWZpY2F0ZQ==\n-----END CERTIFICATE-----"},
					},
				},
			},
		},
		Name: req.Name,
	}, nil
}

func (s *store) UpdateRedisCluster(ctx context.Context, req *clusterpb.UpdateClusterRequest, _ ...gax.CallOption) (gcpclient.ResultOperation[*clusterpb.Cluster], error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	if req.Cluster == nil {
		return nil, gcpmeta.NewBadRequestError("cluster is required")
	}
	if req.UpdateMask == nil || len(req.UpdateMask.Paths) == 0 {
		return nil, gcpmeta.NewBadRequestError("update mask is required")
	}
	if req.Cluster.Name == "" {
		return nil, gcpmeta.NewBadRequestError("cluster name is required")
	}

	rcName, err := gcputil.ParseNameDetail(req.Cluster.Name)
	if err != nil {
		return nil, gcpmeta.NewBadRequestError("invalid cluster name: %v", err)
	}
	if rcName.ResourceType() != gcputil.ResourceTypeCluster {
		return nil, gcpmeta.NewBadRequestError("invalid cluster name type, expected cluster type, got %s", rcName.ResourceType())
	}

	rc, err := s.getRedisClusterNoLock(req.Cluster.Name)
	if err != nil {
		return nil, err
	}

	err = UpdateMask(rc, req.Cluster, req.UpdateMask)
	if err != nil {
		return nil, gcpmeta.NewInternalServerError("update failed: %v", err)
	}
	rc.State = clusterpb.Cluster_UPDATING

	opName := s.newLongRunningOperationName()
	b := NewOperationLongRunningBuilder(opName.String(), rcName)
	if err := b.WithRedisClusterMetadata(rcName, "update"); err != nil {
		return nil, fmt.Errorf("%w: failed setting update redisCluster operation metadata: %w", common.ErrLogical, err)
	}
	s.longRunningOperations.Add(b, opName)

	return NewResultOperation[*clusterpb.Cluster](b.GetOperationPB()), nil
}

func (s *store) DeleteRedisCluster(ctx context.Context, req *clusterpb.DeleteClusterRequest, _ ...gax.CallOption) (gcpclient.VoidOperation, error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	if req.Name == "" {
		return nil, gcpmeta.NewBadRequestError("cluster name is required")
	}
	rcName, err := gcputil.ParseNameDetail(req.Name)
	if err != nil {
		return nil, gcpmeta.NewBadRequestError("invalid cluster name: %v", err)
	}
	if rcName.ResourceType() != gcputil.ResourceTypeCluster {
		return nil, gcpmeta.NewBadRequestError("invalid cluster name type, expected cluster type, got %s", rcName.ResourceType())
	}

	rc, err := s.getRedisClusterNoLock(req.Name)
	if err != nil {
		return nil, err
	}
	rc.State = clusterpb.Cluster_DELETING

	opName := s.newLongRunningOperationName()
	b := NewOperationLongRunningBuilder(opName.String(), rcName)
	if err := b.WithRedisClusterMetadata(rcName, "delete"); err != nil {
		return nil, fmt.Errorf("%w: failed setting delete redisCluster operation metadata: %v", common.ErrLogical, err)
	}
	s.longRunningOperations.Add(b, opName)

	return b.BuildVoidOperation(), nil
}
