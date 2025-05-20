package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/ptr"

	cluster "cloud.google.com/go/redis/cluster/apiv1"
	"cloud.google.com/go/redis/cluster/apiv1/clusterpb"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"

	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

type CreateRedisClusterRequest struct {
	VPCNetworkFullName string
	NodeType           string
	RedisConfigs       map[string]string
	ReplicaCount       int32
	ShardCount         int32
}

type MemorystoreClusterClient interface {
	CreateRedisCluster(ctx context.Context, projectId, locationId, clusterId string, options CreateRedisClusterRequest) error
	GetRedisCluster(ctx context.Context, projectId, locationId, clusterId string) (*clusterpb.Cluster, error)
	GetRedisClusterCertificateAuthority(ctx context.Context, projectId, locationId, clusterId string) (string, error)
	UpdateRedisCluster(ctx context.Context, redisCluster *clusterpb.Cluster, updateMask []string) error
	DeleteRedisCluster(ctx context.Context, projectId, locationId, clusterId string) error
}

func NewMemorystoreClientProvider(gcpClients *gcpclient.GcpClients) gcpclient.GcpClientProvider[MemorystoreClusterClient] {
	return func() MemorystoreClusterClient {
		return NewMemorystoreClient(gcpClients)
	}
}

func NewMemorystoreClient(gcpClients *gcpclient.GcpClients) MemorystoreClusterClient {
	return &memorystoreClient{redisClusterClient: gcpClients.RedisCluster}
}

type memorystoreClient struct {
	redisClusterClient *cluster.CloudRedisClusterClient
}

func (memorystoreClient *memorystoreClient) GetRedisClusterCertificateAuthority(ctx context.Context, projectId, locationId, clusterId string) (string, error) {
	name := GetGcpMemoryStoreRedisClusterName(projectId, locationId, clusterId)
	response, err := memorystoreClient.redisClusterClient.GetClusterCertificateAuthority(ctx, &clusterpb.GetClusterCertificateAuthorityRequest{
		Name: name,
	})

	if err != nil {
		return "", err
	}

	certsWrapped := response.GetManagedServerCa().GetCaCerts()
	if len(certsWrapped) > 0 {
		return strings.Join(certsWrapped[0].Certificates, ""), nil
	}

	return "", nil
}

func (memorystoreClient *memorystoreClient) UpdateRedisCluster(ctx context.Context, redisCluster *clusterpb.Cluster, updateMask []string) error {
	req := &clusterpb.UpdateClusterRequest{
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: updateMask,
		},
		Cluster: redisCluster,
	}

	_, err := memorystoreClient.redisClusterClient.UpdateCluster(ctx, req)
	if err != nil {
		logger := composed.LoggerFromCtx(ctx)
		logger.Error(err, "Failed to update redis cluster", "redisCluster", redisCluster.Name)
		return err
	}

	return nil
}

func (memorystoreClient *memorystoreClient) CreateRedisCluster(ctx context.Context, projectId, locationId, clusterId string, options CreateRedisClusterRequest) error {
	parent := fmt.Sprintf("projects/%s/locations/%s", projectId, locationId)
	req := &clusterpb.CreateClusterRequest{
		Parent:    parent,
		ClusterId: GetGcpMemoryStoreRedisClusterId(clusterId),
		Cluster: &clusterpb.Cluster{
			Name:         GetGcpMemoryStoreRedisClusterName(projectId, locationId, clusterId),
			ReplicaCount: ptr.To(options.ReplicaCount),
			ShardCount:   ptr.To(options.ShardCount),
			NodeType:     clusterpb.NodeType(clusterpb.NodeType_value[options.NodeType]),
			PscConfigs: []*clusterpb.PscConfig{{
				Network: options.VPCNetworkFullName,
			}},
			RedisConfigs: options.RedisConfigs,

			PersistenceConfig:      &clusterpb.ClusterPersistenceConfig{Mode: clusterpb.ClusterPersistenceConfig_DISABLED},
			AuthorizationMode:      clusterpb.AuthorizationMode_AUTH_MODE_DISABLED,
			TransitEncryptionMode:  clusterpb.TransitEncryptionMode_TRANSIT_ENCRYPTION_MODE_SERVER_AUTHENTICATION,
			ZoneDistributionConfig: &clusterpb.ZoneDistributionConfig{Mode: clusterpb.ZoneDistributionConfig_MULTI_ZONE},

			DeletionProtectionEnabled: ptr.To(false),
		},
	}

	_, err := memorystoreClient.redisClusterClient.CreateCluster(ctx, req)

	if err != nil {
		logger := composed.LoggerFromCtx(ctx)
		logger.Error(err, "CreateRedisCluster", "projectId", projectId, "locationId", locationId, "clusterId", clusterId)
		return err
	}

	return nil
}

func (memorystoreClient *memorystoreClient) GetRedisCluster(ctx context.Context, projectId, locationId, clusterId string) (*clusterpb.Cluster, error) {
	logger := composed.LoggerFromCtx(ctx).WithValues("projectId", projectId, "locationId", locationId, "clusterId", clusterId)

	name := GetGcpMemoryStoreRedisClusterName(projectId, locationId, clusterId)
	req := &clusterpb.GetClusterRequest{
		Name: name,
	}

	instanceResponse, err := memorystoreClient.redisClusterClient.GetCluster(ctx, req)
	if err != nil {
		if gcpmeta.IsNotFound(err) {
			logger.Info("target Redis instance not found")
			return nil, err
		}
		logger.Error(err, "Failed to get Redis instance")
		return nil, err
	}

	return instanceResponse, nil
}

func (memorystoreClient *memorystoreClient) DeleteRedisCluster(ctx context.Context, projectId string, locationId string, clusterId string) error {
	req := &clusterpb.DeleteClusterRequest{
		Name: GetGcpMemoryStoreRedisClusterName(projectId, locationId, clusterId),
	}

	_, err := memorystoreClient.redisClusterClient.DeleteCluster(ctx, req)

	if err != nil {
		logger := composed.LoggerFromCtx(ctx)
		logger.Error(err, "DeleteRedisCluster", "projectId", projectId, "locationId", locationId, "clusterId", clusterId)
		return err
	}

	return nil
}
