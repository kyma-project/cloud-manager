package client

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
)

type Client interface {
	azureclient.RedisClient
	azureclient.PrivateEndPointsClient
	azureclient.PrivateDnsZoneGroupClient
}

func NewClientProvider() azureclient.ClientProvider[Client] {
	return func(ctx context.Context, clientId, clientSecret, subscriptionId, tenantId string, auxiliaryTenants ...string) (Client, error) {

		cred, err := azidentity.NewClientSecretCredential(tenantId, clientId, clientSecret, &azidentity.ClientSecretCredentialOptions{})

		if err != nil {
			return nil, err
		}

		armRedisClientInstance, err := armredis.NewClient(subscriptionId, cred, nil)
		if err != nil {
			return nil, err
		}

		privateEndPointsClient, err := armnetwork.NewPrivateEndpointsClient(subscriptionId, cred, nil)
		if err != nil {
			return nil, err
		}

		privateDnsZoneGroupClient, err := armnetwork.NewPrivateDNSZoneGroupsClient(subscriptionId, cred, nil)
		if err != nil {
			return nil, err
		}

		return newClient(
			azureclient.NewRedisClient(armRedisClientInstance),
			azureclient.NewPrivateEndPointClient(privateEndPointsClient),
			azureclient.NewPrivateDnsZoneGroupClient(privateDnsZoneGroupClient),
		), nil
	}
}

type redisInstanceClient struct {
	azureclient.RedisClient
	azureclient.PrivateEndPointsClient
	azureclient.PrivateDnsZoneGroupClient
}

func newClient(redisClient azureclient.RedisClient, privateEndPointsClient azureclient.PrivateEndPointsClient, privateDnsZoneGroupClient azureclient.PrivateDnsZoneGroupClient) Client {
	return &redisInstanceClient{
		RedisClient:               redisClient,
		PrivateEndPointsClient:    privateEndPointsClient,
		PrivateDnsZoneGroupClient: privateDnsZoneGroupClient,
	}
}
