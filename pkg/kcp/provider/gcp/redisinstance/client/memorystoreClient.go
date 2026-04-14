package client

import (
	"context"
	"fmt"

	"cloud.google.com/go/redis/apiv1/redispb"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
)

type CreateRedisInstanceOptions struct {
	VPCNetworkFullName string
	IPRangeName        string
	MemorySizeGb       int32
	Tier               string
	RedisVersion       string
	AuthEnabled        bool
	RedisConfigs       map[string]string
	MaintenancePolicy  *cloudcontrolv1beta1.MaintenancePolicyGcp
	Labels             map[string]string
	ReplicaCount       int32
}

// MemorystoreClient embeds the wrapped gcpclient.RedisInstanceClient interface and adds
// value-add methods that contain real business logic.
// Actions call the wrapped methods directly for simple operations (e.g., UpdateRedisInstance,
// UpgradeRedisInstance, DeleteRedisInstance) using name-building utilities from util.go.
type MemorystoreClient interface {
	gcpclient.RedisInstanceClient

	// CreateRedisInstanceWithOptions creates a Redis instance with complex options-to-protobuf conversion.
	CreateRedisInstanceWithOptions(ctx context.Context, projectId, locationId, instanceId string, options CreateRedisInstanceOptions) error
	// GetRedisInstanceWithAuth gets a Redis instance and conditionally fetches auth string based on AuthEnabled flag.
	GetRedisInstanceWithAuth(ctx context.Context, projectId, locationId, instanceId string) (*redispb.Instance, *redispb.InstanceAuthString, error)
}

func NewMemorystoreClientProvider(gcpClients *gcpclient.GcpClients) gcpclient.GcpClientProvider[MemorystoreClient] {
	return func(_ string) MemorystoreClient {
		return NewMemorystoreClientFromRedisInstanceClient(gcpClients.RedisInstanceWrapped())
	}
}

// NewMemorystoreClientFromRedisInstanceClient wraps a RedisInstanceClient into a MemorystoreClient.
// Cannot be eliminated because MemorystoreClient has value-add methods (CreateRedisInstanceWithOptions,
// GetRedisInstanceWithAuth) beyond the embedded gcpclient.RedisInstanceClient, so a plain
// RedisInstanceClient does not satisfy the interface.
func NewMemorystoreClientFromRedisInstanceClient(redisInstanceClient gcpclient.RedisInstanceClient) MemorystoreClient {
	return &memorystoreClient{
		RedisInstanceClient: redisInstanceClient,
	}
}

type memorystoreClient struct {
	gcpclient.RedisInstanceClient
}

var _ MemorystoreClient = &memorystoreClient{}

func (c *memorystoreClient) CreateRedisInstanceWithOptions(ctx context.Context, projectId, locationId, instanceId string, options CreateRedisInstanceOptions) error {
	readReplicasMode := redispb.Instance_READ_REPLICAS_DISABLED
	if options.Tier != "BASIC" {
		readReplicasMode = redispb.Instance_READ_REPLICAS_ENABLED
	}

	parent := fmt.Sprintf("projects/%s/locations/%s", projectId, locationId)
	req := &redispb.CreateInstanceRequest{
		Parent:     parent,
		InstanceId: GetGcpMemoryStoreRedisInstanceId(instanceId),
		Instance: &redispb.Instance{
			Name:                  GetGcpMemoryStoreRedisName(projectId, locationId, instanceId),
			MemorySizeGb:          options.MemorySizeGb,
			Tier:                  redispb.Instance_Tier(redispb.Instance_Tier_value[options.Tier]),
			RedisVersion:          options.RedisVersion,
			ConnectMode:           redispb.Instance_PRIVATE_SERVICE_ACCESS, // always
			AuthorizedNetwork:     options.VPCNetworkFullName,
			ReservedIpRange:       options.IPRangeName,
			RedisConfigs:          options.RedisConfigs,
			AuthEnabled:           options.AuthEnabled,
			TransitEncryptionMode: redispb.Instance_SERVER_AUTHENTICATION,
			MaintenancePolicy:     ToMaintenancePolicy(options.MaintenancePolicy),
			Labels:                options.Labels,
			ReplicaCount:          options.ReplicaCount,
			ReadReplicasMode:      readReplicasMode,
		},
	}

	_, err := c.CreateRedisInstance(ctx, req)
	return err
}

func (c *memorystoreClient) GetRedisInstanceWithAuth(ctx context.Context, projectId, locationId, instanceId string) (*redispb.Instance, *redispb.InstanceAuthString, error) {
	name := GetGcpMemoryStoreRedisName(projectId, locationId, instanceId)

	instanceResponse, err := c.GetRedisInstance(ctx, &redispb.GetInstanceRequest{Name: name})
	if err != nil {
		return nil, nil, err
	}

	if !instanceResponse.AuthEnabled {
		return instanceResponse, nil, nil
	}

	authResponse, err := c.GetRedisInstanceAuthString(ctx, &redispb.GetInstanceAuthStringRequest{Name: name})
	if err != nil {
		return nil, nil, err
	}

	return instanceResponse, authResponse, nil
}
