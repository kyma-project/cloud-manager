package mock

import (
	"context"
	"sync"

	aws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	elasticacheTypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/redisinstance/client"
	"k8s.io/utils/ptr"
)

type AwsElastiCacheMockUtils interface {
	GetAwsElastiCacheByName(name string) *elasticacheTypes.CacheCluster
	SetAwsElastiCacheLifeCycleState(name string, state awsmeta.ElastiCacheState)
	DeleteAwsElastiCacheByName(name string)
}

type elastiCacheClientFake struct {
	subnetGroupMutex    *sync.Mutex
	parameterGroupMutex *sync.Mutex
	elasticacheMutex    *sync.Mutex
	elastiCaches        map[string]*elasticacheTypes.CacheCluster
	parameterGroups     map[string]*elasticacheTypes.CacheParameterGroup
	subnetGroups        map[string]*elasticacheTypes.CacheSubnetGroup
}

func (client *elastiCacheClientFake) GetAwsElastiCacheByName(name string) *elasticacheTypes.CacheCluster {
	return client.elastiCaches[name]
}

func (client *elastiCacheClientFake) SetAwsElastiCacheLifeCycleState(name string, state awsmeta.ElastiCacheState) {
	if instance, ok := client.elastiCaches[name]; ok {
		instance.CacheClusterStatus = ptr.To(state)
	}
}

func (client *elastiCacheClientFake) DeleteAwsElastiCacheByName(name string) {
	client.elasticacheMutex.Lock()
	defer client.elasticacheMutex.Unlock()

	delete(client.elastiCaches, name)
}

func (client *elastiCacheClientFake) DescribeElastiCacheSubnetGroup(ctx context.Context, name string) ([]elasticacheTypes.CacheSubnetGroup, error) {
	client.subnetGroupMutex.Lock()
	defer client.subnetGroupMutex.Unlock()

	subnetGroup := client.subnetGroups[name]

	if subnetGroup == nil {
		return []elasticacheTypes.CacheSubnetGroup{}, nil
	}

	return []elasticacheTypes.CacheSubnetGroup{*subnetGroup}, nil
}

func (client *elastiCacheClientFake) CreateElastiCacheSubnetGroup(ctx context.Context, name string, subnetIds []string, tags []elasticacheTypes.Tag) (*elasticache.CreateCacheSubnetGroupOutput, error) {
	client.subnetGroupMutex.Lock()
	defer client.subnetGroupMutex.Unlock()

	client.subnetGroups[name] = &elasticacheTypes.CacheSubnetGroup{
		CacheSubnetGroupName: ptr.To(name),
	}

	return &elasticache.CreateCacheSubnetGroupOutput{CacheSubnetGroup: &elasticacheTypes.CacheSubnetGroup{
		CacheSubnetGroupName: ptr.To(name),
	}}, nil
}

func (client *elastiCacheClientFake) DeleteElastiCacheSubnetGroup(ctx context.Context, name string) error {
	client.subnetGroupMutex.Lock()
	defer client.subnetGroupMutex.Unlock()

	delete(client.subnetGroups, name)

	return nil
}

func (client *elastiCacheClientFake) DescribeElastiCacheParameterGroup(ctx context.Context, name string) ([]elasticacheTypes.CacheParameterGroup, error) {
	client.parameterGroupMutex.Lock()
	defer client.parameterGroupMutex.Unlock()

	parameterGroup := client.parameterGroups[name]

	if parameterGroup == nil {
		return []elasticacheTypes.CacheParameterGroup{}, nil
	}

	return []elasticacheTypes.CacheParameterGroup{*parameterGroup}, nil
}

func (client *elastiCacheClientFake) CreateElastiCacheParameterGroup(ctx context.Context, name, family string, tags []elasticacheTypes.Tag) (*elasticache.CreateCacheParameterGroupOutput, error) {
	client.parameterGroupMutex.Lock()
	defer client.parameterGroupMutex.Unlock()

	client.parameterGroups[name] = &elasticacheTypes.CacheParameterGroup{
		CacheParameterGroupName:   ptr.To(name),
		CacheParameterGroupFamily: ptr.To(family),
	}

	return &elasticache.CreateCacheParameterGroupOutput{
		CacheParameterGroup: client.parameterGroups[name],
	}, nil
}

func (client *elastiCacheClientFake) DeleteElastiCacheParameterGroup(ctx context.Context, name string) error {
	client.parameterGroupMutex.Lock()
	defer client.parameterGroupMutex.Unlock()

	delete(client.parameterGroups, name)

	return nil
}

func (client *elastiCacheClientFake) DescribeElastiCacheCluster(ctx context.Context, clusterId string) ([]elasticacheTypes.CacheCluster, error) {
	client.elasticacheMutex.Lock()
	defer client.elasticacheMutex.Unlock()

	cacheCluster := client.elastiCaches[clusterId]

	if cacheCluster == nil {
		return []elasticacheTypes.CacheCluster{}, nil
	}

	return []elasticacheTypes.CacheCluster{*cacheCluster}, nil
}

func (client *elastiCacheClientFake) CreateElastiCacheCluster(ctx context.Context, tags []elasticacheTypes.Tag, options awsclient.CreateElastiCacheClusterOptions) (*elasticache.CreateCacheClusterOutput, error) {
	client.elasticacheMutex.Lock()
	defer client.elasticacheMutex.Unlock()

	client.elastiCaches[options.Name] = &elasticacheTypes.CacheCluster{
		CacheClusterId:       ptr.To(options.Name),
		CacheSubnetGroupName: ptr.To(options.SubnetGroupName),
		CacheClusterStatus:   ptr.To("creating"),
		CacheNodes: []elasticacheTypes.CacheNode{
			{
				Endpoint: &elasticacheTypes.Endpoint{
					Address: ptr.To("192.168.3.3"),
					Port:    aws.Int32(6949),
				},
			},
		},
	}

	return &elasticache.CreateCacheClusterOutput{}, nil
}

func (client *elastiCacheClientFake) DeleteElastiCacheClaster(ctx context.Context, id string) error {
	client.elasticacheMutex.Lock()
	defer client.elasticacheMutex.Unlock()

	if instance, ok := client.elastiCaches[id]; ok {
		instance.CacheClusterStatus = ptr.To("deleting")
	}

	return nil
}
