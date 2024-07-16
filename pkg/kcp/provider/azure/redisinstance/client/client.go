package client

import (
	redis "cloud.google.com/go/redis/apiv1"
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	armRedis "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureClient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
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

type Client interface {
	CreateRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string, parameters armRedis.CreateParameters) (*redis.CreateInstanceOperation, error)
	GetRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string) (*armRedis.ResourceInfo, error)
	DeleteRedisInstance(ctx context.Context, projectId, locationId, instanceId string) error
}

func NewClientProvider() azureClient.SkrClientProvider[Client] {
	return func(ctx context.Context, clientId, clientSecret, subscriptionId, tenantId string) (Client, error) {

		cred, err := azidentity.NewClientSecretCredential(tenantId, clientId, clientSecret, &azidentity.ClientSecretCredentialOptions{})

		if err != nil {
			return nil, err
		}

		armRedisClientInstance, err := armRedis.NewClient(subscriptionId, cred, nil)

		if err != nil {
			return nil, err
		}

		return newClient(armRedisClientInstance), nil
	}
}

type redisClient struct {
	RedisClient *armRedis.Client
}

func newClient(armRedisClientInstance *armRedis.Client) Client {
	return &redisClient{
		RedisClient: armRedisClientInstance,
	}
}

func (c *redisClient) CreateRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string, parameters armRedis.CreateParameters) (*redis.CreateInstanceOperation, error) {
	logger := composed.LoggerFromCtx(ctx)
	_, error := c.RedisClient.BeginCreate(
		ctx,
		resourceGroupName,
		redisInstanceName,
		parameters,
		nil)

	if error != nil {
		logger.Error(error, "Failed to create Azure Redis instance")
		return nil, error
	}
	return nil, nil
}

func (c *redisClient) GetRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string) (*armRedis.ResourceInfo, error) {
	logger := composed.LoggerFromCtx(ctx)

	clientGetResponse, error := c.RedisClient.Get(ctx, resourceGroupName, redisInstanceName, nil)
	if error != nil {
		logger.Error(error, "Failed to get Azure Redis instance")
		return nil, error
	}
	return &clientGetResponse.ResourceInfo, nil
}
func (c *redisClient) DeleteRedisInstance(ctx context.Context, projectId, locationId, instanceId string) error {
	return nil
}
