package client

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redisenterprise/armredisenterprise/v3"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
)

// ManagedRedisClient defines operations for Azure Managed Redis (Microsoft.Cache/redisEnterprise).
type ManagedRedisClient interface {
	// CreateOrUpdateCluster creates or updates an Azure Managed Redis cluster.
	CreateOrUpdateCluster(ctx context.Context, resourceGroupName, clusterName string, cluster armredisenterprise.Cluster) error
	// UpdateCluster performs a partial update (PATCH) of an Azure Managed Redis cluster.
	UpdateCluster(ctx context.Context, resourceGroupName, clusterName string, cluster armredisenterprise.ClusterUpdate) error
	// GetCluster retrieves the current state of an Azure Managed Redis cluster.
	GetCluster(ctx context.Context, resourceGroupName, clusterName string) (*armredisenterprise.Cluster, error)
	// DeleteCluster begins deletion of an Azure Managed Redis cluster.
	DeleteCluster(ctx context.Context, resourceGroupName, clusterName string) error
	// CreateOrUpdateDatabase creates or updates a database within an Azure Managed Redis cluster.
	CreateOrUpdateDatabase(ctx context.Context, resourceGroupName, clusterName, databaseName string, db armredisenterprise.Database) error
	// UpdateDatabase performs a partial update (PATCH) of a database within a cluster.
	UpdateDatabase(ctx context.Context, resourceGroupName, clusterName, databaseName string, db armredisenterprise.DatabaseUpdate) error
	// GetDatabase retrieves the current state of a database within a cluster.
	GetDatabase(ctx context.Context, resourceGroupName, clusterName, databaseName string) (*armredisenterprise.Database, error)
	// DeleteDatabase begins deletion of a database within a cluster.
	DeleteDatabase(ctx context.Context, resourceGroupName, clusterName, databaseName string) error
	// ListKeys retrieves the access keys for the default database.
	ListKeys(ctx context.Context, resourceGroupName, clusterName, databaseName string) (*armredisenterprise.AccessKeys, error)
	// ListSKUsForScaling returns the SKUs that the cluster can be scaled to.
	ListSKUsForScaling(ctx context.Context, resourceGroupName, clusterName string) ([]string, error)
}

// Client composes ManagedRedisClient with reused networking clients.
type Client interface {
	ManagedRedisClient
	azureclient.PrivateEndPointsClient
	azureclient.PrivateDnsZoneGroupClient
	azureclient.PrivateDnsZoneClient
	azureclient.VirtualNetworkLinkClient
}

// NewClientProvider returns a factory that creates a Client from Azure credentials.
func NewClientProvider() azureclient.ClientProvider[Client] {
	return func(ctx context.Context, clientId, clientSecret, subscriptionId, tenantId string, auxiliaryTenants ...string) (Client, error) {
		cred, err := azidentity.NewClientSecretCredential(tenantId, clientId, clientSecret, azureclient.NewCredentialOptionsBuilder().Build())
		if err != nil {
			return nil, err
		}

		redisEnterpriseClient, err := armredisenterprise.NewClient(subscriptionId, cred, azureclient.NewClientOptionsBuilder().Build())
		if err != nil {
			return nil, err
		}

		databasesClient, err := armredisenterprise.NewDatabasesClient(subscriptionId, cred, azureclient.NewClientOptionsBuilder().Build())
		if err != nil {
			return nil, err
		}

		privateEndPointsClient, err := armnetwork.NewPrivateEndpointsClient(subscriptionId, cred, azureclient.NewClientOptionsBuilder().Build())
		if err != nil {
			return nil, err
		}

		privateDnsZoneGroupClient, err := armnetwork.NewPrivateDNSZoneGroupsClient(subscriptionId, cred, azureclient.NewClientOptionsBuilder().Build())
		if err != nil {
			return nil, err
		}

		privateDnsClientFactory, err := armprivatedns.NewClientFactory(subscriptionId, cred, azureclient.NewClientOptionsBuilder().Build())
		if err != nil {
			return nil, err
		}

		return newClient(
			newManagedRedisClient(redisEnterpriseClient, databasesClient),
			azureclient.NewPrivateEndPointClient(privateEndPointsClient),
			azureclient.NewPrivateDnsZoneGroupClient(privateDnsZoneGroupClient),
			azureclient.NewPrivateDnsZoneClient(privateDnsClientFactory.NewPrivateZonesClient()),
			azureclient.NewVirtualNetworkLinkClient(privateDnsClientFactory.NewVirtualNetworkLinksClient()),
		), nil
	}
}

type managedRedisClientImpl struct {
	clustersClient  *armredisenterprise.Client
	databasesClient *armredisenterprise.DatabasesClient
}

func newManagedRedisClient(clustersClient *armredisenterprise.Client, databasesClient *armredisenterprise.DatabasesClient) ManagedRedisClient {
	return &managedRedisClientImpl{
		clustersClient:  clustersClient,
		databasesClient: databasesClient,
	}
}

func (c *managedRedisClientImpl) CreateOrUpdateCluster(ctx context.Context, resourceGroupName, clusterName string, cluster armredisenterprise.Cluster) error {
	_, err := c.clustersClient.BeginCreate(ctx, resourceGroupName, clusterName, cluster, nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *managedRedisClientImpl) UpdateCluster(ctx context.Context, resourceGroupName, clusterName string, cluster armredisenterprise.ClusterUpdate) error {
	_, err := c.clustersClient.BeginUpdate(ctx, resourceGroupName, clusterName, cluster, nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *managedRedisClientImpl) GetCluster(ctx context.Context, resourceGroupName, clusterName string) (*armredisenterprise.Cluster, error) {
	resp, err := c.clustersClient.Get(ctx, resourceGroupName, clusterName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Cluster, nil
}

func (c *managedRedisClientImpl) DeleteCluster(ctx context.Context, resourceGroupName, clusterName string) error {
	_, err := c.clustersClient.BeginDelete(ctx, resourceGroupName, clusterName, nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *managedRedisClientImpl) CreateOrUpdateDatabase(ctx context.Context, resourceGroupName, clusterName, databaseName string, db armredisenterprise.Database) error {
	_, err := c.databasesClient.BeginCreate(ctx, resourceGroupName, clusterName, databaseName, db, nil)
	return err
}

func (c *managedRedisClientImpl) UpdateDatabase(ctx context.Context, resourceGroupName, clusterName, databaseName string, db armredisenterprise.DatabaseUpdate) error {
	_, err := c.databasesClient.BeginUpdate(ctx, resourceGroupName, clusterName, databaseName, db, nil)
	return err
}

func (c *managedRedisClientImpl) GetDatabase(ctx context.Context, resourceGroupName, clusterName, databaseName string) (*armredisenterprise.Database, error) {
	resp, err := c.databasesClient.Get(ctx, resourceGroupName, clusterName, databaseName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Database, nil
}

func (c *managedRedisClientImpl) DeleteDatabase(ctx context.Context, resourceGroupName, clusterName, databaseName string) error {
	_, err := c.databasesClient.BeginDelete(ctx, resourceGroupName, clusterName, databaseName, nil)
	return err
}

func (c *managedRedisClientImpl) ListKeys(ctx context.Context, resourceGroupName, clusterName, databaseName string) (*armredisenterprise.AccessKeys, error) {
	resp, err := c.databasesClient.ListKeys(ctx, resourceGroupName, clusterName, databaseName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.AccessKeys, nil
}

func (c *managedRedisClientImpl) ListSKUsForScaling(ctx context.Context, resourceGroupName, clusterName string) ([]string, error) {
	resp, err := c.clustersClient.ListSKUsForScaling(ctx, resourceGroupName, clusterName, nil)
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(resp.SKUs))
	for _, sku := range resp.SKUs {
		if sku != nil && sku.Name != nil {
			result = append(result, *sku.Name)
		}
	}
	return result, nil
}

// compositeClient embeds all sub-clients.
type compositeClient struct {
	ManagedRedisClient
	azureclient.PrivateEndPointsClient
	azureclient.PrivateDnsZoneGroupClient
	azureclient.PrivateDnsZoneClient
	azureclient.VirtualNetworkLinkClient
}

func newClient(
	managedRedis ManagedRedisClient,
	peClient azureclient.PrivateEndPointsClient,
	dnsGroupClient azureclient.PrivateDnsZoneGroupClient,
	dnsZoneClient azureclient.PrivateDnsZoneClient,
	vnetLinkClient azureclient.VirtualNetworkLinkClient,
) Client {
	return &compositeClient{
		ManagedRedisClient:        managedRedis,
		PrivateEndPointsClient:    peClient,
		PrivateDnsZoneGroupClient: dnsGroupClient,
		PrivateDnsZoneClient:      dnsZoneClient,
		VirtualNetworkLinkClient:  vnetLinkClient,
	}
}
