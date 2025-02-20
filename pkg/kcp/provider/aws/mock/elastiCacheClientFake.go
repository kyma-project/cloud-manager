package mock

import (
	"context"
	"fmt"
	"sync"

	aws "github.com/aws/aws-sdk-go-v2/aws"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	elasticacheTypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	secretsmanager "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	secretsmanagerTypes "github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/elliotchance/pie/v2"
	"github.com/google/uuid"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"k8s.io/utils/ptr"
)

type AwsElastiCacheMockUtils interface {
	GetAwsElastiCacheByName(name string) *elasticacheTypes.ReplicationGroup
	GetAWsElastiCacheNodeByName(name string) *elasticacheTypes.CacheCluster
	SetAwsElastiCacheLifeCycleState(name string, state awsmeta.ElastiCacheState)
	SetAwsElastiCacheEngineVersion(name, engineVersion string)
	SetAwsElastiCacheUserGroupLifeCycleState(name string, state awsmeta.ElastiCacheUserGroupState)
	DeleteAwsElastiCacheByName(name string)
	DeleteAwsElastiCacheUserGroupByName(name string)
	DescribeAwsElastiCacheParametersByName(groupName string) map[string]string
}

func getDefaultParams() map[string]elasticacheTypes.Parameter {
	return map[string]elasticacheTypes.Parameter{
		"maxmemory-policy": {
			ParameterName:  ptr.To("maxmemory-policy"),
			ParameterValue: ptr.To("volatile-lru"),
		},
		"active-defrag-cycle-max": {
			ParameterName:  ptr.To("active-defrag-cycle-max"),
			ParameterValue: ptr.To("75"),
		},
		"acl-pubsub-default": {
			ParameterName:  ptr.To("acl-pubsub-default"),
			ParameterValue: ptr.To("allchannels"),
		},
		"cluster-enabled": {
			ParameterName:  ptr.To("cluster-enabled"),
			ParameterValue: ptr.To("no"),
		},
	}
}

type elastiCacheClientFake struct {
	mutex             sync.Mutex
	replicationGroups map[string]*elasticacheTypes.ReplicationGroup
	cacheClusters     map[string]*elasticacheTypes.CacheCluster
	parameters        map[string]map[string]elasticacheTypes.Parameter
	parameterGroups   map[string]*elasticacheTypes.CacheParameterGroup
	subnetGroups      map[string]*elasticacheTypes.CacheSubnetGroup
	userGroups        map[string]*elasticacheTypes.UserGroup
	secretStore       map[string]*secretsmanager.GetSecretValueOutput
	securityGroups    []*ec2Types.SecurityGroup
}

func newElastiCacheClientFake() *elastiCacheClientFake {
	return &elastiCacheClientFake{
		mutex:             sync.Mutex{},
		replicationGroups: map[string]*elasticacheTypes.ReplicationGroup{},
		cacheClusters:     map[string]*elasticacheTypes.CacheCluster{},
		subnetGroups:      map[string]*elasticacheTypes.CacheSubnetGroup{},
		parameterGroups:   map[string]*elasticacheTypes.CacheParameterGroup{},
		parameters:        map[string]map[string]elasticacheTypes.Parameter{},
		secretStore:       map[string]*secretsmanager.GetSecretValueOutput{},
		userGroups:        map[string]*elasticacheTypes.UserGroup{},
		securityGroups:    []*ec2Types.SecurityGroup{},
	}
}

func (client *elastiCacheClientFake) GetAwsElastiCacheByName(name string) *elasticacheTypes.ReplicationGroup {
	return client.replicationGroups[name]
}

func (client *elastiCacheClientFake) GetAWsElastiCacheNodeByName(name string) *elasticacheTypes.CacheCluster {
	return client.cacheClusters[name]
}

func (client *elastiCacheClientFake) SetAwsElastiCacheLifeCycleState(name string, state awsmeta.ElastiCacheState) {
	if instance, ok := client.replicationGroups[name]; ok {
		instance.Status = ptr.To(state)
	}
}

func (client *elastiCacheClientFake) SetAwsElastiCacheEngineVersion(name, engineVersion string) {
	if instance, ok := client.cacheClusters[name]; ok {
		instance.EngineVersion = ptr.To(engineVersion)
		instance.PendingModifiedValues = nil
	}
}

func (client *elastiCacheClientFake) SetAwsElastiCacheUserGroupLifeCycleState(name string, state awsmeta.ElastiCacheUserGroupState) {
	if instance, ok := client.userGroups[name]; ok {
		instance.Status = ptr.To(state)
	}
}

func (client *elastiCacheClientFake) DeleteAwsElastiCacheByName(name string) {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	delete(client.replicationGroups, name)
}

func (client *elastiCacheClientFake) DeleteAwsElastiCacheUserGroupByName(name string) {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	delete(client.userGroups, name)
}

func (client *elastiCacheClientFake) DescribeAwsElastiCacheParametersByName(groupName string) map[string]string {
	result := map[string]string{}

	for _, parameter := range pie.Values(client.parameters[groupName]) {
		result[*parameter.ParameterName] = *parameter.ParameterValue
	}

	return result
}

func (client *elastiCacheClientFake) DescribeElastiCacheSubnetGroup(ctx context.Context, name string) ([]elasticacheTypes.CacheSubnetGroup, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	client.mutex.Lock()
	defer client.mutex.Unlock()

	subnetGroup := client.subnetGroups[name]

	if subnetGroup == nil {
		return []elasticacheTypes.CacheSubnetGroup{}, nil
	}

	return []elasticacheTypes.CacheSubnetGroup{*subnetGroup}, nil
}

func (client *elastiCacheClientFake) CreateElastiCacheSubnetGroup(ctx context.Context, name string, subnetIds []string, tags []elasticacheTypes.Tag) (*elasticache.CreateCacheSubnetGroupOutput, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	client.mutex.Lock()
	defer client.mutex.Unlock()

	client.subnetGroups[name] = &elasticacheTypes.CacheSubnetGroup{
		CacheSubnetGroupName: ptr.To(name),
	}

	return &elasticache.CreateCacheSubnetGroupOutput{CacheSubnetGroup: &elasticacheTypes.CacheSubnetGroup{
		CacheSubnetGroupName: ptr.To(name),
	}}, nil
}

func (client *elastiCacheClientFake) DeleteElastiCacheSubnetGroup(ctx context.Context, name string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}

	client.mutex.Lock()
	defer client.mutex.Unlock()

	delete(client.subnetGroups, name)

	return nil
}

func (client *elastiCacheClientFake) DescribeElastiCacheParameterGroup(ctx context.Context, name string) ([]elasticacheTypes.CacheParameterGroup, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	client.mutex.Lock()
	defer client.mutex.Unlock()

	parameterGroup := client.parameterGroups[name]

	if parameterGroup == nil {
		return []elasticacheTypes.CacheParameterGroup{}, nil
	}

	return []elasticacheTypes.CacheParameterGroup{*parameterGroup}, nil
}

func (client *elastiCacheClientFake) CreateElastiCacheParameterGroup(ctx context.Context, name, family string, tags []elasticacheTypes.Tag) (*elasticache.CreateCacheParameterGroupOutput, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	client.mutex.Lock()
	defer client.mutex.Unlock()

	client.parameterGroups[name] = &elasticacheTypes.CacheParameterGroup{
		CacheParameterGroupName:   ptr.To(name),
		CacheParameterGroupFamily: ptr.To(family),
	}

	client.parameters[name] = getDefaultParams()

	return &elasticache.CreateCacheParameterGroupOutput{
		CacheParameterGroup: client.parameterGroups[name],
	}, nil
}

func (client *elastiCacheClientFake) DeleteElastiCacheParameterGroup(ctx context.Context, name string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}

	client.mutex.Lock()
	defer client.mutex.Unlock()

	delete(client.parameterGroups, name)

	return nil
}

func (client *elastiCacheClientFake) DescribeElastiCacheParameters(ctx context.Context, groupName string) ([]elasticacheTypes.Parameter, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	client.mutex.Lock()
	defer client.mutex.Unlock()

	return pie.Values(client.parameters[groupName]), nil
}

func (client *elastiCacheClientFake) ModifyElastiCacheParameterGroup(ctx context.Context, groupName string, parameters []elasticacheTypes.ParameterNameValue) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}

	client.mutex.Lock()
	defer client.mutex.Unlock()

	for _, parameter := range parameters {
		client.parameters[groupName][ptr.Deref(parameter.ParameterName, "")] = elasticacheTypes.Parameter{ParameterName: parameter.ParameterName, ParameterValue: parameter.ParameterValue}
	}

	return nil
}

func (client *elastiCacheClientFake) DescribeEngineDefaultParameters(ctx context.Context, family string) ([]elasticacheTypes.Parameter, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	return pie.Values(getDefaultParams()), nil
}

func (client *elastiCacheClientFake) GetAuthTokenSecretValue(ctx context.Context, secretName string) (*secretsmanager.GetSecretValueOutput, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	client.mutex.Lock()
	defer client.mutex.Unlock()

	return client.secretStore[secretName], nil
}

func (client *elastiCacheClientFake) CreateAuthTokenSecret(ctx context.Context, secretName string, tags []secretsmanagerTypes.Tag) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}

	client.mutex.Lock()
	defer client.mutex.Unlock()

	client.secretStore[secretName] = &secretsmanager.GetSecretValueOutput{
		Name:         ptr.To(secretName),
		SecretString: ptr.To(uuid.NewString()),
	}

	return nil
}

func (client *elastiCacheClientFake) DeleteAuthTokenSecret(ctx context.Context, secretName string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}

	client.mutex.Lock()
	defer client.mutex.Unlock()

	delete(client.secretStore, secretName)

	return nil
}

func (client *elastiCacheClientFake) DescribeElastiCacheReplicationGroup(ctx context.Context, clusterId string) ([]elasticacheTypes.ReplicationGroup, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	client.mutex.Lock()
	defer client.mutex.Unlock()

	cacheCluster := client.replicationGroups[clusterId]

	if cacheCluster == nil {
		return []elasticacheTypes.ReplicationGroup{}, nil
	}

	return []elasticacheTypes.ReplicationGroup{*cacheCluster}, nil
}

func (client *elastiCacheClientFake) CreateElastiCacheReplicationGroup(ctx context.Context, tags []elasticacheTypes.Tag, options awsclient.CreateElastiCacheClusterOptions) (*elasticache.CreateReplicationGroupOutput, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	client.mutex.Lock()
	defer client.mutex.Unlock()

	client.cacheClusters[options.Name] = &elasticacheTypes.CacheCluster{
		CacheClusterId:             ptr.To(options.Name),
		PreferredMaintenanceWindow: options.PreferredMaintenanceWindow,
		CacheParameterGroup: &elasticacheTypes.CacheParameterGroupStatus{
			CacheParameterGroupName: &options.ParameterGroupName,
		},
		EngineVersion: &options.EngineVersion,
		Engine:        ptr.To("redis"),
	}

	authTokenEnabled := false
	if options.AuthTokenSecretString != nil {
		authTokenEnabled = true
	}

	client.replicationGroups[options.Name] = &elasticacheTypes.ReplicationGroup{
		ReplicationGroupId:       ptr.To(options.Name),
		Status:                   ptr.To("creating"),
		CacheNodeType:            ptr.To(options.CacheNodeType),
		AutoMinorVersionUpgrade:  ptr.To(options.AutoMinorVersionUpgrade),
		TransitEncryptionEnabled: ptr.To(true),
		AuthTokenEnabled:         ptr.To(authTokenEnabled),
		MemberClusters:           []string{options.Name},
		UserGroupIds:             []string{},
		Engine:                   ptr.To("redis"),
		AtRestEncryptionEnabled:  ptr.To(true),
		ConfigurationEndpoint: &elasticacheTypes.Endpoint{
			Address: ptr.To("192.168.3.2"),
			Port:    aws.Int32(6949),
		},
		NodeGroups: []elasticacheTypes.NodeGroup{
			{
				PrimaryEndpoint: &elasticacheTypes.Endpoint{
					Address: ptr.To("192.168.3.3"),
					Port:    aws.Int32(6949),
				},
				ReaderEndpoint: &elasticacheTypes.Endpoint{
					Address: ptr.To("192.168.3.4"),
					Port:    aws.Int32(6949),
				},
			},
		},
	}

	return &elasticache.CreateReplicationGroupOutput{}, nil
}

func (client *elastiCacheClientFake) ModifyElastiCacheReplicationGroup(ctx context.Context, id string, options awsclient.ModifyElastiCacheClusterOptions) (*elasticache.ModifyReplicationGroupOutput, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	client.mutex.Lock()
	defer client.mutex.Unlock()

	if instance, ok := client.replicationGroups[id]; ok {
		instance.Status = ptr.To("modifying")
		if options.CacheNodeType != nil {
			instance.CacheNodeType = options.CacheNodeType
		}
		if options.AutoMinorVersionUpgrade != nil {
			instance.AutoMinorVersionUpgrade = options.AutoMinorVersionUpgrade
		}

		if len(options.UserGroupIdsToAdd) > 0 {
			instance.UserGroupIds = append(instance.UserGroupIds, options.UserGroupIdsToAdd...)
			instance.AuthTokenEnabled = ptr.To(false)
		}
		if len(options.UserGroupIdsToRemove) > 0 {
			_, remaining := pie.Diff(instance.UserGroupIds, options.UserGroupIdsToRemove)
			instance.UserGroupIds = remaining
		}

		if options.AuthTokenSecretString != nil {
			instance.AuthTokenEnabled = ptr.To(true)
		}
	}

	if instance, ok := client.cacheClusters[id]; ok {
		if options.ParameterGroupName != nil {
			instance.CacheParameterGroup = &elasticacheTypes.CacheParameterGroupStatus{
				CacheParameterGroupName: options.ParameterGroupName,
			}
		}

		if options.EngineVersion != nil {
			instance.PendingModifiedValues = &elasticacheTypes.PendingModifiedValues{
				EngineVersion: options.EngineVersion,
			}
		}
	}

	return &elasticache.ModifyReplicationGroupOutput{}, nil
}

func (client *elastiCacheClientFake) DeleteElastiCacheReplicationGroup(ctx context.Context, id string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}

	client.mutex.Lock()
	defer client.mutex.Unlock()

	if instance, ok := client.replicationGroups[id]; ok {
		instance.Status = ptr.To("deleting")
	}

	return nil
}

func (client *elastiCacheClientFake) DescribeElastiCacheCluster(ctx context.Context, id string) ([]elasticacheTypes.CacheCluster, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	client.mutex.Lock()
	defer client.mutex.Unlock()

	cacheCluster := client.cacheClusters[id]

	if cacheCluster == nil {
		return []elasticacheTypes.CacheCluster{}, nil
	}

	return []elasticacheTypes.CacheCluster{*cacheCluster}, nil
}

func (client *elastiCacheClientFake) DescribeUserGroup(ctx context.Context, id string) (*elasticacheTypes.UserGroup, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	client.mutex.Lock()
	defer client.mutex.Unlock()

	userGroup := client.userGroups[id]

	return userGroup, nil
}

func (client *elastiCacheClientFake) CreateUserGroup(ctx context.Context, id string, tags []elasticacheTypes.Tag) (*elasticache.CreateUserGroupOutput, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	client.mutex.Lock()
	defer client.mutex.Unlock()

	client.userGroups[id] = &elasticacheTypes.UserGroup{
		Engine:      ptr.To("redis"),
		UserGroupId: ptr.To(id),
		Status:      ptr.To("creating"),
		UserIds:     []string{"default"},
	}

	return &elasticache.CreateUserGroupOutput{UserGroupId: ptr.To(id)}, nil
}

func (client *elastiCacheClientFake) DeleteUserGroup(ctx context.Context, id string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}

	client.mutex.Lock()
	defer client.mutex.Unlock()

	if instance, ok := client.userGroups[id]; ok {
		instance.Status = ptr.To("deleting")
	}

	return nil
}

func (client *elastiCacheClientFake) DescribeElastiCacheSecurityGroups(ctx context.Context, filters []ec2Types.Filter, groupIds []string) ([]ec2Types.SecurityGroup, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	client.mutex.Lock()
	defer client.mutex.Unlock()

	list := append([]*ec2Types.SecurityGroup{}, client.securityGroups...)
	if groupIds != nil {
		list = pie.Filter(list, func(sg *ec2Types.SecurityGroup) bool {
			return pie.Contains(groupIds, ptr.Deref(sg.GroupId, ""))
		})
	}
	if filters != nil {
		list = pie.Filter(list, func(sg *ec2Types.SecurityGroup) bool {
			return anyFilterMatchTags(sg.Tags, filters)
		})
	}
	result := make([]ec2Types.SecurityGroup, 0, len(list))
	for _, x := range list {
		result = append(result, *x)
	}
	return result, nil
}

func (client *elastiCacheClientFake) CreateElastiCacheSecurityGroup(ctx context.Context, vpcId string, name string, tags []ec2Types.Tag) (string, error) {
	if isContextCanceled(ctx) {
		return "", context.Canceled
	}

	client.mutex.Lock()
	defer client.mutex.Unlock()

	tags = append(tags, ec2Types.Tag{
		Key:   ptr.To("vpc-id"),
		Value: ptr.To(vpcId),
	})
	sg := &ec2Types.SecurityGroup{
		Description: ptr.To(name),
		GroupId:     ptr.To(uuid.NewString()),
		GroupName:   ptr.To(name),
		Tags:        tags,
		VpcId:       ptr.To(vpcId),
	}
	client.securityGroups = append(client.securityGroups, sg)
	return ptr.Deref(sg.GroupId, ""), nil
}

func (client *elastiCacheClientFake) AuthorizeElastiCacheSecurityGroupIngress(ctx context.Context, groupId string, ipPermissions []ec2Types.IpPermission) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}

	client.mutex.Lock()
	defer client.mutex.Unlock()

	var securityGroup *ec2Types.SecurityGroup
	for _, sg := range client.securityGroups {
		if ptr.Deref(sg.GroupId, "") == groupId {
			securityGroup = sg
			break
		}
	}
	if securityGroup == nil {
		return fmt.Errorf("security group with id %s does not exist", groupId)
	}
	securityGroup.IpPermissions = ipPermissions
	return nil
}

func (client *elastiCacheClientFake) DeleteElastiCacheSecurityGroup(ctx context.Context, id string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}

	client.mutex.Lock()
	defer client.mutex.Unlock()

	client.securityGroups = pie.Filter(client.securityGroups, func(sg *ec2Types.SecurityGroup) bool {
		return ptr.Deref(sg.GroupId, "") != id
	})
	return nil
}
