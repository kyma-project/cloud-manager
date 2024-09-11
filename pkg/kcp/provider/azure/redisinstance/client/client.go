package client

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	azureClient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	"k8s.io/utils/ptr"
)

type Client interface {
	CreateRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string, parameters armredis.CreateParameters) error
	UpdateRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string, parameters armredis.UpdateParameters) error
	GetRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string) (*armredis.ResourceInfo, error)
	DeleteRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string) error
	GetRedisInstanceAccessKeys(ctx context.Context, resourceGroupName, redisInstanceName string) ([]string, error)
	//GetResourceGroup(ctx context.Context, name string) (*armresources.ResourceGroupsClientGetResponse, error)
	//CreateResourceGroup(ctx context.Context, name string, location string) error
	//DeleteResourceGroup(ctx context.Context, name string) error
}

func NewClientProvider() azureClient.SkrClientProvider[Client] {
	return func(ctx context.Context, clientId, clientSecret, subscriptionId, tenantId string) (Client, error) {

		cred, err := azidentity.NewClientSecretCredential(tenantId, clientId, clientSecret, &azidentity.ClientSecretCredentialOptions{})

		if err != nil {
			return nil, err
		}

		armRedisClientInstance, err := armredis.NewClient(subscriptionId, cred, nil)

		if err != nil {
			return nil, err
		}

		resourceGroupClientInstance, err := armresources.NewResourceGroupsClient(subscriptionId, cred, nil)

		if err != nil {
			return nil, err
		}

		return newClient(armRedisClientInstance, resourceGroupClientInstance), nil
	}
}

type redisClient struct {
	RedisClient         *armredis.Client
	ResourceGroupClient *armresources.ResourceGroupsClient
}

func newClient(armRedisClientInstance *armredis.Client, resourceGroupClientInstance *armresources.ResourceGroupsClient) Client {
	return &redisClient{
		RedisClient:         armRedisClientInstance,
		ResourceGroupClient: resourceGroupClientInstance,
	}
}

func (c *redisClient) CreateRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string, parameters armredis.CreateParameters) error {
	_, err := c.RedisClient.BeginCreate(
		ctx,
		resourceGroupName,
		redisInstanceName,
		parameters,
		nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *redisClient) GetRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string) (*armredis.ResourceInfo, error) {
	clientGetResponse, err := c.RedisClient.Get(ctx, resourceGroupName, redisInstanceName, nil)
	if err != nil {
		return nil, err
	}
	return &clientGetResponse.ResourceInfo, nil
}
func (c *redisClient) DeleteRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string) error {
	_, err := c.RedisClient.BeginDelete(ctx, resourceGroupName, redisInstanceName, nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *redisClient) GetRedisInstanceAccessKeys(ctx context.Context, resourceGroupName, redisInstanceName string) ([]string, error) {
	redisAccessKeys, err := c.RedisClient.ListKeys(ctx, resourceGroupName, redisInstanceName, nil)

	if err != nil {
		return nil, err
	}
	return []string{ptr.Deref(redisAccessKeys.PrimaryKey, "")}, nil
}

//func (c *redisClient) GetResourceGroup(ctx context.Context, name string) (*armresources.ResourceGroupsClientGetResponse, error) {
//	resourceGroupsClientGetResponse, err := c.ResourceGroupClient.Get(ctx, name, nil)
//
//	if err != nil {
//		return nil, err
//	}
//
//	return &resourceGroupsClientGetResponse, nil
//}
//
//func (c *redisClient) CreateResourceGroup(ctx context.Context, name string, location string) error {
//	resourceGroup := armresources.ResourceGroup{Location: to.Ptr(location)}
//	_, err := c.ResourceGroupClient.CreateOrUpdate(ctx, name, resourceGroup, nil)
//
//	if err != nil {
//		return err
//	}
//
//	return nil
//}
//func (c *redisClient) DeleteResourceGroup(ctx context.Context, name string) error {
//	_, err := c.ResourceGroupClient.BeginDelete(ctx, name, nil)
//
//	if err != nil {
//		return err
//	}
//
//	return nil
//}

func (c *redisClient) UpdateRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string, parameters armredis.UpdateParameters) error {
	_, err := c.RedisClient.Update(
		ctx,
		resourceGroupName,
		redisInstanceName,
		parameters,
		nil)

	if err != nil {
		return err
	}

	return nil
}
