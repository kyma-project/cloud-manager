package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/ptr"

	cluster "cloud.google.com/go/redis/cluster/apiv1"
	"cloud.google.com/go/redis/cluster/apiv1/clusterpb"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
	"google.golang.org/api/option"
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

func NewMemorystoreClientProvider() client.ClientProvider[MemorystoreClusterClient] {
	return func(ctx context.Context, saJsonKeyPath string) (MemorystoreClusterClient, error) {
		return NewMemorystoreClient(saJsonKeyPath), nil
	}
}

func NewMemorystoreClient(saJsonKeyPath string) MemorystoreClusterClient {
	return &memorystoreClient{saJsonKeyPath: saJsonKeyPath}
}

type memorystoreClient struct {
	saJsonKeyPath string
}

func (memorystoreClient *memorystoreClient) GetRedisClusterCertificateAuthority(ctx context.Context, projectId, locationId, clusterId string) (string, error) {
	redisClient, err := cluster.NewCloudRedisClusterClient(ctx, option.WithCredentialsFile(memorystoreClient.saJsonKeyPath))
	if err != nil {
		return "", err
	}
	defer redisClient.Close() // nolint: errcheck

	name := GetGcpMemoryStoreRedisClusterName(projectId, locationId, clusterId)
	response, err := redisClient.GetClusterCertificateAuthority(ctx, &clusterpb.GetClusterCertificateAuthorityRequest{
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
	redisClient, err := cluster.NewCloudRedisClusterClient(ctx, option.WithCredentialsFile(memorystoreClient.saJsonKeyPath))
	if err != nil {
		return err
	}
	defer redisClient.Close() // nolint: errcheck
	req := &clusterpb.UpdateClusterRequest{
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: updateMask,
		},
		Cluster: redisCluster,
	}

	_, err = redisClient.UpdateCluster(ctx, req)
	if err != nil {
		logger := composed.LoggerFromCtx(ctx)
		logger.Error(err, "Failed to update redis cluster", "redisCluster", redisCluster.Name)
		return err
	}

	return nil
}

func (memorystoreClient *memorystoreClient) CreateRedisCluster(ctx context.Context, projectId, locationId, clusterId string, options CreateRedisClusterRequest) error {
	redisClient, err := cluster.NewCloudRedisClusterClient(ctx, option.WithCredentialsFile(memorystoreClient.saJsonKeyPath))
	if err != nil {
		return err
	}
	defer redisClient.Close() // nolint: errcheck

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

	_, err = redisClient.CreateCluster(ctx, req)

	if err != nil {
		logger := composed.LoggerFromCtx(ctx)
		logger.Error(err, "CreateRedisCluster", "projectId", projectId, "locationId", locationId, "clusterId", clusterId)
		return err
	}

	return nil
}

func (memorystoreClient *memorystoreClient) GetRedisCluster(ctx context.Context, projectId, locationId, clusterId string) (*clusterpb.Cluster, error) {
	logger := composed.LoggerFromCtx(ctx).WithValues("projectId", projectId, "locationId", locationId, "clusterId", clusterId)
	redisClient, err := cluster.NewCloudRedisClusterClient(ctx, option.WithCredentialsFile(memorystoreClient.saJsonKeyPath))
	if err != nil {
		return nil, err
	}
	defer redisClient.Close() // nolint: errcheck

	name := GetGcpMemoryStoreRedisClusterName(projectId, locationId, clusterId)
	req := &clusterpb.GetClusterRequest{
		Name: name,
	}

	instanceResponse, err := redisClient.GetCluster(ctx, req)
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
	redisClient, redisClientErr := cluster.NewCloudRedisClusterClient(ctx, option.WithCredentialsFile(memorystoreClient.saJsonKeyPath))
	if redisClientErr != nil {
		return redisClientErr
	}
	defer redisClient.Close() // nolint: errcheck

	req := &clusterpb.DeleteClusterRequest{
		Name: GetGcpMemoryStoreRedisClusterName(projectId, locationId, clusterId),
	}

	_, err := redisClient.DeleteCluster(ctx, req)

	if err != nil {
		logger := composed.LoggerFromCtx(ctx)
		logger.Error(err, "DeleteRedisCluster", "projectId", projectId, "locationId", locationId, "clusterId", clusterId)
		return err
	}

	return nil
}
