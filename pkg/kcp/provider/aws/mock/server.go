package mock

import (
	"context"
	"sync"

	elasticacheTypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	secretsmanager "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	iprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/iprange/client"
	nfsinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/nfsinstance/client"
	redisinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/redisinstance/client"
	vpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/vpcpeering/client"
	scopeclient "github.com/kyma-project/cloud-manager/pkg/kcp/scope/client"
)

var _ Server = &server{}

func New() Server {
	return &server{
		vpcStore:         &vpcStore{},
		nfsStore:         &nfsStore{},
		scopeStore:       &scopeStore{},
		vpcPeeringStore:  &vpcPeeringStore{},
		routeTablesStore: &routeTablesStore{},
		elastiCacheClientFake: &elastiCacheClientFake{
			elasticacheMutex:    &sync.Mutex{},
			subnetGroupMutex:    &sync.Mutex{},
			parameterGroupMutex: &sync.Mutex{},
			secretStoreMutex:    &sync.Mutex{},
			userGroupsMutex:     &sync.Mutex{},
			replicationGroups:   map[string]*elasticacheTypes.ReplicationGroup{},
			cacheClusters:       map[string]*elasticacheTypes.CacheCluster{},
			subnetGroups:        map[string]*elasticacheTypes.CacheSubnetGroup{},
			parameterGroups:     map[string]*elasticacheTypes.CacheParameterGroup{},
			parameters:          map[string]map[string]elasticacheTypes.Parameter{},
			secretStore:         map[string]*secretsmanager.GetSecretValueOutput{},
			userGroups:          map[string]*elasticacheTypes.UserGroup{},
		},
	}
}

type server struct {
	*vpcStore
	*nfsStore
	*scopeStore
	*vpcPeeringStore
	*elastiCacheClientFake
	*routeTablesStore
}

func (s *server) ScopeGardenProvider() awsclient.GardenClientProvider[scopeclient.AwsStsClient] {
	return func(ctx context.Context, region, key, secret string) (scopeclient.AwsStsClient, error) {
		return s, nil
	}
}

func (s *server) IpRangeSkrProvider() awsclient.SkrClientProvider[iprangeclient.Client] {
	return func(ctx context.Context, region, key, secret, role string) (iprangeclient.Client, error) {
		return s, nil
	}
}

func (s *server) NfsInstanceSkrProvider() awsclient.SkrClientProvider[nfsinstanceclient.Client] {
	return func(ctx context.Context, region, key, secret, role string) (nfsinstanceclient.Client, error) {
		return s, nil
	}
}

func (s *server) VpcPeeringSkrProvider() awsclient.SkrClientProvider[vpcpeeringclient.Client] {
	return func(ctx context.Context, region, key, secret, role string) (vpcpeeringclient.Client, error) {
		return s, nil
	}
}

func (s *server) ElastiCacheProviderFake() awsclient.SkrClientProvider[redisinstanceclient.ElastiCacheClient] {
	return func(ctx context.Context, region, key, secret, role string) (redisinstanceclient.ElastiCacheClient, error) {
		return s, nil
	}
}
