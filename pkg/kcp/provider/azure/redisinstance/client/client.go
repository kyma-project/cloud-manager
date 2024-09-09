package client

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	armRedis "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis"
	armResources "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureClient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
)

type Client interface {
	CreateRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string, parameters armRedis.CreateParameters) error
	UpdateRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string, parameters armRedis.UpdateParameters) error
	GetRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string) (*armRedis.ResourceInfo, error)
	DeleteRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string) error
	GetRedisInstanceAccessKeys(ctx context.Context, resourceGroupName, redisInstanceName string) (string, error)
	GetResourceGroup(ctx context.Context, name string) (*armResources.ResourceGroupsClientGetResponse, error)
	CreateResourceGroup(ctx context.Context, name string, location string) error
	DeleteResourceGroup(ctx context.Context, name string) error
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

		resourceGroupClientInstance, err := armResources.NewResourceGroupsClient(subscriptionId, cred, nil)

		if err != nil {
			return nil, err
		}

		return newClient(armRedisClientInstance, resourceGroupClientInstance), nil
	}
}

type redisClient struct {
	RedisClient         *armRedis.Client
	ResourceGroupClient *armResources.ResourceGroupsClient
}

func newClient(armRedisClientInstance *armRedis.Client, resourceGroupClientInstance *armResources.ResourceGroupsClient) Client {
	return &redisClient{
		RedisClient:         armRedisClientInstance,
		ResourceGroupClient: resourceGroupClientInstance,
	}
}

func (c *redisClient) CreateRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string, parameters armRedis.CreateParameters) error {
	logger := composed.LoggerFromCtx(ctx)
	_, err := c.RedisClient.BeginCreate(
		ctx,
		resourceGroupName,
		redisInstanceName,
		parameters,
		nil)

	if err != nil {
		logger.Error(err, "Failed to create Azure Redis instance")
		return err
	}

	return nil
}

func (c *redisClient) GetRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string) (*armRedis.ResourceInfo, error) {
	logger := composed.LoggerFromCtx(ctx)

	clientGetResponse, err := c.RedisClient.Get(ctx, resourceGroupName, redisInstanceName, nil)
	if err != nil {
		logger.Error(err, "Failed to get Azure Redis instance")
		return nil, err
	}
	return &clientGetResponse.ResourceInfo, nil
}
func (c *redisClient) DeleteRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string) error {
	logger := composed.LoggerFromCtx(ctx)

	_, err := c.RedisClient.BeginDelete(ctx, resourceGroupName, redisInstanceName, nil)
	if err != nil {
		logger.Error(err, "Failed to delete Azure Redis instance")
		return err
	}
	return nil
}

func (c *redisClient) GetRedisInstanceAccessKeys(ctx context.Context, resourceGroupName, redisInstanceName string) (string, error) {
	logger := composed.LoggerFromCtx(ctx)

	redisAccessKeys, err := c.RedisClient.ListKeys(ctx, resourceGroupName, redisInstanceName, nil)

	if err != nil {
		logger.Error(err, "Failed to get Azure Redis access keys")
		return "", err
	}
	return *redisAccessKeys.PrimaryKey, nil
}

func (c *redisClient) GetResourceGroup(ctx context.Context, name string) (*armResources.ResourceGroupsClientGetResponse, error) {
	logger := composed.LoggerFromCtx(ctx)

	resourceGroupsClientGetResponse, err := c.ResourceGroupClient.Get(ctx, name, nil)

	if err != nil {
		logger.Error(err, "Failed to get Azure Redis resource group")
		return nil, err
	}

	return &resourceGroupsClientGetResponse, nil
}

func (c *redisClient) CreateResourceGroup(ctx context.Context, name string, location string) error {
	logger := composed.LoggerFromCtx(ctx)

	resourceGroup := armResources.ResourceGroup{Location: to.Ptr(location)}
	_, err := c.ResourceGroupClient.CreateOrUpdate(ctx, name, resourceGroup, nil)

	if err != nil {
		logger.Error(err, "Failed to create Azure Redis resource group")
		return err
	}

	return nil
}
func (c *redisClient) DeleteResourceGroup(ctx context.Context, name string) error {
	logger := composed.LoggerFromCtx(ctx)

	_, err := c.ResourceGroupClient.BeginDelete(ctx, name, nil)

	if err != nil {
		logger.Error(err, "Failed to delete Azure Redis resource group")
		return err
	}

	return nil
}

func (c *redisClient) UpdateRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string, parameters armRedis.UpdateParameters) error {
	logger := composed.LoggerFromCtx(ctx)
	_, err := c.RedisClient.Update(
		ctx,
		resourceGroupName,
		redisInstanceName,
		parameters,
		nil)

	if err != nil {
		logger.Error(err, "Failed to update Azure Redis instance")
		return err
	}

	return nil
}
