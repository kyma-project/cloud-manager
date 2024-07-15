package client

import (
	redis "cloud.google.com/go/redis/apiv1"
	redispb "cloud.google.com/go/redis/apiv1/redispb"
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	armRedis "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis"
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
	CreateRedisInstance(ctx context.Context, projectId, locationId, instanceId string, options CreateRedisInstanceOptions) (*redis.CreateInstanceOperation, error)
	GetRedisInstance(ctx context.Context, projectId, locationId, instanceId string) (*redispb.Instance, *redispb.InstanceAuthString, error)
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

type memorystoreClient struct {
}

func newClient(armRedisClientInstance *armRedis.Client) Client {
	return &memorystoreClient{}
}

func (memorystoreClient *memorystoreClient) CreateRedisInstance(ctx context.Context, projectId, locationId, instanceId string, options CreateRedisInstanceOptions) (*redis.CreateInstanceOperation, error) {
	return nil, nil
}

func (memorystoreClient *memorystoreClient) GetRedisInstance(ctx context.Context, projectId, locationId, instanceId string) (*redispb.Instance, *redispb.InstanceAuthString, error) {
	return nil, nil, nil
}
func (memorystoreClient *memorystoreClient) DeleteRedisInstance(ctx context.Context, projectId, locationId, instanceId string) error {
	return nil
}
