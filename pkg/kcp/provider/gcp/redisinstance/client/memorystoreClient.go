package client

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/composed"

	redis "cloud.google.com/go/redis/apiv1"
	redispb "cloud.google.com/go/redis/apiv1/redispb"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"google.golang.org/api/option"
)

type CreateRedisInstanceOptions struct {
	VPCNetworkFullName    string
	IPRangeName           string
	MemorySizeGb          int32
	Tier                  string
	RedisVersion          string
	AuthEnabled           bool
	TransitEncryptionMode string
	RedisConfigs          map[string]string
}

type MemorystoreClient interface {
	CreateRedisInstance(ctx context.Context, projectId, locationId, instanceId string, options CreateRedisInstanceOptions) (*redis.CreateInstanceOperation, error)
	GetRedisInstance(ctx context.Context, projectId, locationId, instanceId string) (*redispb.Instance, *redispb.InstanceAuthString, error)
	DeleteRedisInstance(ctx context.Context, projectId, locationId, instanceId string) error
}

func NewMemorystoreClientProvider() client.ClientProvider[MemorystoreClient] {
	return func(ctx context.Context, saJsonKeyPath string) (MemorystoreClient, error) {
		return NewMemorystoreClient(saJsonKeyPath), nil
	}
}

func NewMemorystoreClient(saJsonKeyPath string) MemorystoreClient {
	return &memorystoreClient{saJsonKeyPath: saJsonKeyPath}
}

type memorystoreClient struct {
	saJsonKeyPath string
}

func (memorystoreClient *memorystoreClient) CreateRedisInstance(ctx context.Context, projectId, locationId, instanceId string, options CreateRedisInstanceOptions) (*redis.CreateInstanceOperation, error) {
	redisClient, err := redis.NewCloudRedisClient(ctx, option.WithCredentialsFile(memorystoreClient.saJsonKeyPath))
	if err != nil {
		return nil, err
	}
	defer redisClient.Close()

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
			TransitEncryptionMode: redispb.Instance_TransitEncryptionMode(redispb.Instance_TransitEncryptionMode_value[options.TransitEncryptionMode]),
		},
	}

	operation, err := redisClient.CreateInstance(ctx, req)

	if err != nil {
		logger := composed.LoggerFromCtx(ctx)
		logger.Error(err, "CreateRedisInstance", "projectId", projectId, "locationId", locationId, "instanceId", instanceId)
		return nil, err
	}

	return operation, nil
}

func (memorystoreClient *memorystoreClient) GetRedisInstance(ctx context.Context, projectId, locationId, instanceId string) (*redispb.Instance, *redispb.InstanceAuthString, error) {
	redisClient, err := redis.NewCloudRedisClient(ctx, option.WithCredentialsFile(memorystoreClient.saJsonKeyPath))
	if err != nil {
		return nil, nil, err
	}
	defer redisClient.Close()

	name := GetGcpMemoryStoreRedisName(projectId, locationId, instanceId)
	req := &redispb.GetInstanceRequest{
		Name: name,
	}

	instanceResponse, err := redisClient.GetInstance(ctx, req)
	if err != nil {
		logger := composed.LoggerFromCtx(ctx)
		logger.Error(err, "Failed to get Redis instance")
		return nil, nil, err
	}

	if !instanceResponse.AuthEnabled {
		return instanceResponse, nil, err
	}

	authResponse, err := redisClient.GetInstanceAuthString(ctx, &redispb.GetInstanceAuthStringRequest{Name: name})
	if err != nil {
		logger := composed.LoggerFromCtx(ctx)
		logger.Error(err, "Failed to get Redis instance Auth")
		return nil, nil, err
	}

	return instanceResponse, authResponse, nil
}

func (memorystoreClient *memorystoreClient) DeleteRedisInstance(ctx context.Context, projectId string, locationId string, instanceId string) error {
	redisClient, redisClientErr := redis.NewCloudRedisClient(ctx, option.WithCredentialsFile(memorystoreClient.saJsonKeyPath))
	if redisClientErr != nil {
		return redisClientErr
	}
	defer redisClient.Close()

	req := &redispb.DeleteInstanceRequest{
		Name: GetGcpMemoryStoreRedisName(projectId, locationId, instanceId),
	}

	_, err := redisClient.DeleteInstance(ctx, req)

	if err != nil {
		logger := composed.LoggerFromCtx(ctx)
		logger.Error(err, "Failed to delete Redis instance")

		return err
	}

	return nil
}
