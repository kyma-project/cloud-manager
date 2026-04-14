package client

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/utils/ptr"

	"cloud.google.com/go/redis/cluster/apiv1/clusterpb"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
)

type CreateRedisClusterRequest struct {
	VPCNetworkFullName string
	NodeType           string
	RedisConfigs       map[string]string
	ReplicaCount       int32
	ShardCount         int32
}

// MemorystoreClusterClient embeds the wrapped gcpclient.RedisClusterClient interface and adds
// value-add methods that contain real business logic.
// Actions call the wrapped methods directly for simple operations (e.g., UpdateRedisCluster,
// DeleteRedisCluster, GetRedisCluster) using name-building utilities from util.go.
type MemorystoreClusterClient interface {
	gcpclient.RedisClusterClient

	// CreateRedisClusterWithOptions creates a Redis cluster with complex options-to-protobuf conversion.
	CreateRedisClusterWithOptions(ctx context.Context, projectId, locationId, clusterId string, options CreateRedisClusterRequest) error
	// GetRedisClusterCertificateString retrieves and extracts the certificate string from nested CA certs structure.
	GetRedisClusterCertificateString(ctx context.Context, projectId, locationId, clusterId string) (string, error)
}

func NewMemorystoreClientProvider(gcpClients *gcpclient.GcpClients) gcpclient.GcpClientProvider[MemorystoreClusterClient] {
	return func(_ string) MemorystoreClusterClient {
		return NewMemorystoreClient(gcpClients)
	}
}

func NewMemorystoreClient(gcpClients *gcpclient.GcpClients) MemorystoreClusterClient {
	return NewMemorystoreClientFromRedisClusterClient(gcpClients.RedisClusterWrapped())
}

func NewMemorystoreClientFromRedisClusterClient(redisClusterClient gcpclient.RedisClusterClient) MemorystoreClusterClient {
	return &memorystoreClient{
		RedisClusterClient: redisClusterClient,
	}
}

type memorystoreClient struct {
	gcpclient.RedisClusterClient
}

var _ MemorystoreClusterClient = &memorystoreClient{}

func (c *memorystoreClient) GetRedisClusterCertificateString(ctx context.Context, projectId, locationId, clusterId string) (string, error) {
	name := GetGcpMemoryStoreRedisClusterName(projectId, locationId, clusterId)
	response, err := c.GetRedisClusterCertificateAuthority(ctx, &clusterpb.GetClusterCertificateAuthorityRequest{
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

func (c *memorystoreClient) CreateRedisClusterWithOptions(ctx context.Context, projectId, locationId, clusterId string, options CreateRedisClusterRequest) error {
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

	_, err := c.CreateRedisCluster(ctx, req)
	return err
}
