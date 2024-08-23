package client

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
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
	GetSubnet(ctx context.Context, resourceGroupName, virtualNetworkName, subnetName string) (*armnetwork.SubnetsClientGetResponse, error)
	CreateSubnet(ctx context.Context, resourceGroupName, virtualNetworkName, subnetName, cidr string) error
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

		subnetClientInstance, err := armnetwork.NewSubnetsClient(subscriptionId, cred, nil)

		if err != nil {
			return nil, err
		}

		return newClient(armRedisClientInstance, resourceGroupClientInstance, subnetClientInstance), nil
	}
}

type redisClient struct {
	RedisClient         *armRedis.Client
	ResourceGroupClient *armResources.ResourceGroupsClient
	SubnetClient        *armnetwork.SubnetsClient
}

func newClient(armRedisClientInstance *armRedis.Client, resourceGroupClientInstance *armResources.ResourceGroupsClient, subnetClientInstance *armnetwork.SubnetsClient) Client {
	return &redisClient{
		RedisClient:         armRedisClientInstance,
		ResourceGroupClient: resourceGroupClientInstance,
		SubnetClient:        subnetClientInstance,
	}
}

func (c *redisClient) CreateRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string, parameters armRedis.CreateParameters) error {
	logger := composed.LoggerFromCtx(ctx)
	_, error := c.RedisClient.BeginCreate(
		ctx,
		resourceGroupName,
		redisInstanceName,
		parameters,
		nil)

	if error != nil {
		logger.Error(error, "Failed to create Azure Redis instance")
		return error
	}

	return nil
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
func (c *redisClient) DeleteRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string) error {
	logger := composed.LoggerFromCtx(ctx)

	_, error := c.RedisClient.BeginDelete(ctx, resourceGroupName, redisInstanceName, nil)
	if error != nil {
		logger.Error(error, "Failed to delete Azure Redis instance")
		return error
	}
	return nil
}

func (c *redisClient) GetRedisInstanceAccessKeys(ctx context.Context, resourceGroupName, redisInstanceName string) (string, error) {
	logger := composed.LoggerFromCtx(ctx)

	redisAccessKeys, error := c.RedisClient.ListKeys(ctx, resourceGroupName, redisInstanceName, nil)

	if error != nil {
		logger.Error(error, "Failed to get Azure Redis access keys")
		return "", error
	}
	return *redisAccessKeys.PrimaryKey, nil
}

func (c *redisClient) GetResourceGroup(ctx context.Context, name string) (*armResources.ResourceGroupsClientGetResponse, error) {
	logger := composed.LoggerFromCtx(ctx)

	resourceGroupsClientGetResponse, error := c.ResourceGroupClient.Get(ctx, name, nil)

	if error != nil {
		logger.Error(error, "Failed to get Azure Redis resource group")
		return nil, error
	}

	return &resourceGroupsClientGetResponse, nil
}

func (c *redisClient) CreateResourceGroup(ctx context.Context, name string, location string) error {
	logger := composed.LoggerFromCtx(ctx)

	resourceGroup := armResources.ResourceGroup{Location: to.Ptr(location)}
	_, error := c.ResourceGroupClient.CreateOrUpdate(ctx, name, resourceGroup, nil)

	if error != nil {
		logger.Error(error, "Failed to create Azure Redis resource group")
		return error
	}

	return nil
}
func (c *redisClient) DeleteResourceGroup(ctx context.Context, name string) error {
	logger := composed.LoggerFromCtx(ctx)

	_, error := c.ResourceGroupClient.BeginDelete(ctx, name, nil)

	if error != nil {
		logger.Error(error, "Failed to delete Azure Redis resource group")
		return error
	}

	return nil
}

func (c *redisClient) UpdateRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string, parameters armRedis.UpdateParameters) error {
	logger := composed.LoggerFromCtx(ctx)
	_, error := c.RedisClient.Update(
		ctx,
		resourceGroupName,
		redisInstanceName,
		parameters,
		nil)

	if error != nil {
		logger.Error(error, "Failed to update Azure Redis instance")
		return error
	}

	return nil
}

func (c *redisClient) GetSubnet(ctx context.Context, resourceGroupName, virtualNetworkName, subnetName string) (*armnetwork.SubnetsClientGetResponse, error) {
	logger := composed.LoggerFromCtx(ctx)

	subnetClientGetResponse, error := c.SubnetClient.Get(ctx, resourceGroupName, virtualNetworkName, subnetName, nil)

	if error != nil {
		logger.Error(error, "Failed to get Azure Redis subnet")
		return nil, error
	}

	return &subnetClientGetResponse, nil
}

func (c *redisClient) CreateSubnet(ctx context.Context, resourceGroupName, virtualNetworkName, subnetName, cidr string) error {
	logger := composed.LoggerFromCtx(ctx)

	subnet := armnetwork.Subnet{
		Properties: &armnetwork.SubnetPropertiesFormat{
			AddressPrefix: to.Ptr(cidr),
		},
	}
	_, error := c.SubnetClient.BeginCreateOrUpdate(ctx, resourceGroupName, virtualNetworkName, subnetName, subnet, nil)

	if error != nil {
		logger.Error(error, "Failed to create Azure Redis subnet")
		return error
	}

	return nil
}
