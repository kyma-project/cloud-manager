package client

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis"
	"k8s.io/utils/ptr"
)

type RedisClient interface {
	CreateRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string, parameters armredis.CreateParameters) error
	UpdateRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string, parameters armredis.UpdateParameters) error
	GetRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string) (*armredis.ResourceInfo, error)
	DeleteRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string) error
	GetRedisInstanceAccessKeys(ctx context.Context, resourceGroupName, redisInstanceName string) ([]string, error)
}

func NewRedisClient(svc *armredis.Client) RedisClient {
	return &redisClient{svc: svc}
}

var _ RedisClient = &redisClient{}

type redisClient struct {
	svc *armredis.Client
}

func (c *redisClient) CreateRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string, parameters armredis.CreateParameters) error {
	_, err := c.svc.BeginCreate(
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
	clientGetResponse, err := c.svc.Get(ctx, resourceGroupName, redisInstanceName, nil)
	if err != nil {
		return nil, err
	}
	return &clientGetResponse.ResourceInfo, nil
}

func (c *redisClient) DeleteRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string) error {
	_, err := c.svc.BeginDelete(ctx, resourceGroupName, redisInstanceName, nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *redisClient) GetRedisInstanceAccessKeys(ctx context.Context, resourceGroupName, redisInstanceName string) ([]string, error) {
	redisAccessKeys, err := c.svc.ListKeys(ctx, resourceGroupName, redisInstanceName, nil)

	if err != nil {
		return nil, err
	}
	return []string{ptr.Deref(redisAccessKeys.PrimaryKey, "")}, nil
}

func (c *redisClient) UpdateRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string, parameters armredis.UpdateParameters) error {
	_, err := c.svc.Update(
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
