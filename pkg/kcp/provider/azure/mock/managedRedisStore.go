package mock

import (
	"context"
	"fmt"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redisenterprise/armredisenterprise/v3"
	azuremanagedredisclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/managedredis/client"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

var _ azuremanagedredisclient.ManagedRedisClient = &managedRedisStore{}

func newManagedRedisStore(subscription string) *managedRedisStore {
	return &managedRedisStore{
		subscription: subscription,
		items:        map[string]map[string]*armredisenterprise.Cluster{},
		databases:    map[string]map[string]map[string]*armredisenterprise.Database{},
		accessKeys:   map[string]map[string]*armredisenterprise.AccessKeys{},
	}
}

type managedRedisStore struct {
	m sync.Mutex

	subscription string

	// items are resourceGroupName => clusterName => Cluster
	items map[string]map[string]*armredisenterprise.Cluster
	// databases are resourceGroupName => clusterName => databaseName => Database
	databases map[string]map[string]map[string]*armredisenterprise.Database
	// accessKeys are resourceGroupName => clusterName => AccessKeys
	accessKeys map[string]map[string]*armredisenterprise.AccessKeys
}

func (s *managedRedisStore) CreateOrUpdateCluster(ctx context.Context, resourceGroupName, clusterName string, cluster armredisenterprise.Cluster) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	if s.items[resourceGroupName] == nil {
		s.items[resourceGroupName] = map[string]*armredisenterprise.Cluster{}
	}

	existing, exists := s.items[resourceGroupName][clusterName]
	if exists {
		// Reject immutable field changes (SKU is immutable after creation)
		if cluster.SKU != nil && cluster.SKU.Name != nil && existing.SKU != nil && existing.SKU.Name != nil {
			if *cluster.SKU.Name != *existing.SKU.Name {
				return fmt.Errorf("mock: SKU is immutable after creation (current=%s, requested=%s)", *existing.SKU.Name, *cluster.SKU.Name)
			}
		}
		// Reject zone changes (HighAvailability is immutable)
		if len(cluster.Zones) > 0 && len(existing.Zones) > 0 {
			if len(cluster.Zones) != len(existing.Zones) {
				return fmt.Errorf("mock: Zones (HighAvailability) is immutable after creation")
			}
		}
		// Update: only mutable fields (TLSVersion)
		if cluster.Properties != nil {
			if cluster.Properties.MinimumTLSVersion != nil {
				existing.Properties.MinimumTLSVersion = cluster.Properties.MinimumTLSVersion
			}
		}
		provisioningState := armredisenterprise.ProvisioningStateSucceeded
		existing.Properties.ProvisioningState = &provisioningState
	} else {
		// Create new
		region := "unknown"
		if cluster.Location != nil {
			region = *cluster.Location
		}
		clusterID := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Cache/redisEnterprise/%s",
			s.subscription, resourceGroupName, clusterName)
		cluster.ID = &clusterID
		cluster.Name = &clusterName
		if cluster.Properties == nil {
			cluster.Properties = &armredisenterprise.ClusterCreateProperties{}
		}
		clusterProvisioningState := armredisenterprise.ProvisioningStateSucceeded
		cluster.Properties.ProvisioningState = &clusterProvisioningState
		hostName := fmt.Sprintf("%s.%s.redisenterprise.cache.azure.net", clusterName, region)
		cluster.Properties.HostName = &hostName

		cloned, _ := util.JsonClone(&cluster)
		s.items[resourceGroupName][clusterName] = cloned
	}

	// Ensure access keys exist
	if s.accessKeys[resourceGroupName] == nil {
		s.accessKeys[resourceGroupName] = map[string]*armredisenterprise.AccessKeys{}
	}
	if _, hasKeys := s.accessKeys[resourceGroupName][clusterName]; !hasKeys {
		primaryKey := "mock-primary-key-" + clusterName
		secondaryKey := "mock-secondary-key-" + clusterName
		s.accessKeys[resourceGroupName][clusterName] = &armredisenterprise.AccessKeys{
			PrimaryKey:   &primaryKey,
			SecondaryKey: &secondaryKey,
		}
	}

	return nil
}

func (s *managedRedisStore) UpdateCluster(ctx context.Context, resourceGroupName, clusterName string, cluster armredisenterprise.ClusterUpdate) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	if s.items[resourceGroupName] == nil {
		return azuremeta.NewAzureNotFoundError()
	}
	existing, exists := s.items[resourceGroupName][clusterName]
	if !exists {
		return azuremeta.NewAzureNotFoundError()
	}

	if cluster.Properties != nil {
		if cluster.Properties.MinimumTLSVersion != nil {
			existing.Properties.MinimumTLSVersion = cluster.Properties.MinimumTLSVersion
		}
	}

	return nil
}

func (s *managedRedisStore) GetCluster(ctx context.Context, resourceGroupName, clusterName string) (*armredisenterprise.Cluster, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	if s.items[resourceGroupName] == nil {
		return nil, azuremeta.NewAzureNotFoundError()
	}
	cluster, exists := s.items[resourceGroupName][clusterName]
	if !exists {
		return nil, azuremeta.NewAzureNotFoundError()
	}

	res, err := util.JsonClone(cluster)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *managedRedisStore) DeleteCluster(ctx context.Context, resourceGroupName, clusterName string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	if s.items[resourceGroupName] == nil {
		return nil // treat as success per design
	}
	delete(s.items[resourceGroupName], clusterName)

	// Clean up access keys
	if s.accessKeys[resourceGroupName] != nil {
		delete(s.accessKeys[resourceGroupName], clusterName)
	}

	// Clean up databases
	if s.databases[resourceGroupName] != nil {
		delete(s.databases[resourceGroupName], clusterName)
	}

	return nil
}

func (s *managedRedisStore) UpdateDatabase(ctx context.Context, resourceGroupName, clusterName, databaseName string, db armredisenterprise.DatabaseUpdate) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	if s.databases[resourceGroupName] == nil || s.databases[resourceGroupName][clusterName] == nil {
		return azuremeta.NewAzureNotFoundError()
	}
	existing, exists := s.databases[resourceGroupName][clusterName][databaseName]
	if !exists {
		return azuremeta.NewAzureNotFoundError()
	}

	if db.Properties != nil {
		if db.Properties.ClientProtocol != nil {
			existing.Properties.ClientProtocol = db.Properties.ClientProtocol
		}
	}

	return nil
}

func (s *managedRedisStore) CreateOrUpdateDatabase(ctx context.Context, resourceGroupName, clusterName, databaseName string, db armredisenterprise.Database) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	if s.databases[resourceGroupName] == nil {
		s.databases[resourceGroupName] = map[string]map[string]*armredisenterprise.Database{}
	}
	if s.databases[resourceGroupName][clusterName] == nil {
		s.databases[resourceGroupName][clusterName] = map[string]*armredisenterprise.Database{}
	}

	// Reject immutable field changes (ClusteringPolicy is immutable after creation)
	if existing, exists := s.databases[resourceGroupName][clusterName][databaseName]; exists {
		if db.Properties != nil && db.Properties.ClusteringPolicy != nil &&
			existing.Properties != nil && existing.Properties.ClusteringPolicy != nil {
			if *db.Properties.ClusteringPolicy != *existing.Properties.ClusteringPolicy {
				return fmt.Errorf("mock: ClusteringPolicy is immutable after creation (current=%s, requested=%s)",
					*existing.Properties.ClusteringPolicy, *db.Properties.ClusteringPolicy)
			}
		}
	}

	if db.Properties == nil {
		db.Properties = &armredisenterprise.DatabaseCreateProperties{}
	}
	dbProvisioningState := armredisenterprise.ProvisioningStateSucceeded
	db.Properties.ProvisioningState = &dbProvisioningState
	db.Name = &databaseName

	cloned, _ := util.JsonClone(&db)
	s.databases[resourceGroupName][clusterName][databaseName] = cloned
	return nil
}

func (s *managedRedisStore) GetDatabase(ctx context.Context, resourceGroupName, clusterName, databaseName string) (*armredisenterprise.Database, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	if s.databases[resourceGroupName] == nil || s.databases[resourceGroupName][clusterName] == nil {
		return nil, azuremeta.NewAzureNotFoundError()
	}
	db, exists := s.databases[resourceGroupName][clusterName][databaseName]
	if !exists {
		return nil, azuremeta.NewAzureNotFoundError()
	}
	res, err := util.JsonClone(db)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *managedRedisStore) DeleteDatabase(ctx context.Context, resourceGroupName, clusterName, databaseName string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	if s.databases[resourceGroupName] != nil && s.databases[resourceGroupName][clusterName] != nil {
		delete(s.databases[resourceGroupName][clusterName], databaseName)
	}
	return nil
}

func (s *managedRedisStore) ListKeys(ctx context.Context, resourceGroupName, clusterName, databaseName string) (*armredisenterprise.AccessKeys, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	if s.items[resourceGroupName] == nil {
		return nil, azuremeta.NewAzureNotFoundError()
	}
	if _, clusterExists := s.items[resourceGroupName][clusterName]; !clusterExists {
		return nil, azuremeta.NewAzureNotFoundError()
	}
	if s.accessKeys[resourceGroupName] == nil {
		return nil, azuremeta.NewAzureNotFoundError()
	}
	keys, exists := s.accessKeys[resourceGroupName][clusterName]
	if !exists {
		return nil, azuremeta.NewAzureNotFoundError()
	}

	res, err := util.JsonClone(keys)
	if err != nil {
		return nil, err
	}
	return res, nil
}
