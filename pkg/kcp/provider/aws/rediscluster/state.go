package rediscluster

import (
	"context"
	"fmt"
	"strings"

	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elasticacheTypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	secretsmanager "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/elliotchance/pie/v2"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	awsconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/config"
	"github.com/kyma-project/cloud-manager/pkg/kcp/rediscluster/types"
	"k8s.io/utils/ptr"
)

type State struct {
	types.State
	awsClient awsclient.ElastiCacheClient

	subnetGroup                           *elasticacheTypes.CacheSubnetGroup
	parameterGroup                        *elasticacheTypes.CacheParameterGroup
	parameterGroupCurrentParams           []elasticacheTypes.Parameter
	parameterGroupFamilyDefaultParams     []elasticacheTypes.Parameter
	tempParameterGroup                    *elasticacheTypes.CacheParameterGroup
	tempParameterGroupFamilyDefaultParams []elasticacheTypes.Parameter
	tempParameterGroupCurrentParams       []elasticacheTypes.Parameter

	elastiCacheReplicationGroup *elasticacheTypes.ReplicationGroup
	memberClusters              []elasticacheTypes.CacheCluster
	authTokenValue              *secretsmanager.GetSecretValueOutput
	userGroup                   *elasticacheTypes.UserGroup
	securityGroup               *ec2Types.SecurityGroup
	securityGroupId             string

	modifyElastiCacheClusterOptions awsclient.ModifyElastiCacheClusterOptions
	updateMask                      []string
}

type StateFactory interface {
	NewState(ctx context.Context, redisInstace types.State) (*State, error)
}

func NewStateFactory(skrProvider awsclient.SkrClientProvider[awsclient.ElastiCacheClient]) StateFactory {
	return &stateFactory{
		skrProvider: skrProvider,
	}
}

type stateFactory struct {
	skrProvider awsclient.SkrClientProvider[awsclient.ElastiCacheClient]
}

func (f *stateFactory) NewState(ctx context.Context, redisInstace types.State) (*State, error) {
	roleName := fmt.Sprintf("arn:aws:iam::%s:role/%s", redisInstace.Scope().Spec.Scope.Aws.AccountId, awsconfig.AwsConfig.Default.AssumeRoleName)

	logger := composed.LoggerFromCtx(ctx)
	logger.
		WithValues(
			"awsRegion", redisInstace.Scope().Spec.Region,
			"awsRole", roleName,
		).
		Info("Assuming AWS role")

	c, err := f.skrProvider(
		ctx,
		redisInstace.Scope().Spec.Scope.Aws.AccountId,
		redisInstace.Scope().Spec.Region,
		awsconfig.AwsConfig.Default.AccessKeyId,
		awsconfig.AwsConfig.Default.SecretAccessKey,
		roleName,
	)
	if err != nil {
		return nil, err
	}

	return newState(redisInstace, c), nil
}

func newState(redisInstace types.State, elastiCacheClient awsclient.ElastiCacheClient) *State {
	return &State{
		State:                           redisInstace,
		awsClient:                       elastiCacheClient,
		modifyElastiCacheClusterOptions: awsclient.ModifyElastiCacheClusterOptions{},
		updateMask:                      []string{},
	}
}

func (s *State) ShouldUpdateRedisCluster() bool {
	return len(s.updateMask) > 0
}

func (s *State) GetModifyElastiCacheClusterOptions() awsclient.ModifyElastiCacheClusterOptions {
	return s.modifyElastiCacheClusterOptions
}

func (s *State) UpdateCacheNodeType(cacheNodeType string) {
	s.modifyElastiCacheClusterOptions.CacheNodeType = ptr.To(cacheNodeType)
	s.updateMask = append(s.updateMask, "cacheNodeType")
}

func (s *State) UpdateAutoMinorVersionUpgrade(autoMinorVersionUpgrade bool) {
	s.modifyElastiCacheClusterOptions.AutoMinorVersionUpgrade = ptr.To(autoMinorVersionUpgrade)
	s.updateMask = append(s.updateMask, "autoMinorVersionUpgrade")
}

func (s *State) UpdatePreferredMaintenanceWindow(preferredMaintenanceWindow string) {
	s.modifyElastiCacheClusterOptions.PreferredMaintenanceWindow = ptr.To(preferredMaintenanceWindow)
	s.updateMask = append(s.updateMask, "preferredMaintenanceWindow")
}

func (s *State) UpdateAuthEnabled(authEnabled bool) {
	s.updateMask = append(s.updateMask, "authEnabled")
	if authEnabled {
		s.modifyElastiCacheClusterOptions.AuthTokenSecretString = s.authTokenValue.SecretString
	} else {
		if len(s.elastiCacheReplicationGroup.UserGroupIds) < 1 {
			s.modifyElastiCacheClusterOptions.UserGroupIdsToAdd = []string{ptr.Deref(s.userGroup.UserGroupId, "")}
		} else {
			s.modifyElastiCacheClusterOptions.UserGroupIdsToRemove = []string{ptr.Deref(s.userGroup.UserGroupId, "")}
		}
	}

}

func (s *State) AreMainParamGroupParamsUpToDate() bool {
	currentParametersMap := MapParameters(s.parameterGroupCurrentParams)
	defaultParametersMap := MapParameters(s.parameterGroupFamilyDefaultParams)

	desiredParametersMap := GetDesiredParameters(defaultParametersMap, s.ObjAsRedisCluster().Spec.Instance.Aws.Parameters)
	forUpdateParameters := GetMissmatchedParameters(currentParametersMap, desiredParametersMap)

	return len(forUpdateParameters) == 0
}

func (s *State) IsMainParamGroupFamilyUpToDate() bool {
	desiredParamGroupFamily := GetAwsElastiCacheParameterGroupFamily(s.ObjAsRedisCluster().Spec.Instance.Aws.EngineVersion)
	currentParamGroupFamily := ptr.Deref(s.parameterGroup.CacheParameterGroupFamily, "")

	return desiredParamGroupFamily == currentParamGroupFamily
}

func (s *State) AreTempParamGroupParamsUpToDate() bool {
	currentParametersMap := MapParameters(s.tempParameterGroupCurrentParams)
	defaultParametersMap := MapParameters(s.tempParameterGroupFamilyDefaultParams)

	desiredParametersMap := GetDesiredParameters(defaultParametersMap, s.ObjAsRedisCluster().Spec.Instance.Aws.Parameters)
	forUpdateParameters := GetMissmatchedParameters(currentParametersMap, desiredParametersMap)

	return len(forUpdateParameters) == 0
}

func (s *State) IsRedisVersionUpToDate() bool {
	for _, memberCluster := range s.memberClusters {
		if memberCluster.CacheParameterGroup == nil {
			return true
		}

		desiredVersion := s.ObjAsRedisCluster().Spec.Instance.Aws.EngineVersion
		currentVersion := ptr.Deref(memberCluster.EngineVersion, "")

		if strings.HasPrefix(currentVersion, "6") && strings.HasPrefix(strings.ToLower(desiredVersion), "6.x") {
			return true
		}

		if !strings.Contains(currentVersion, desiredVersion) {
			return false
		}
	}

	return true
}

func (s *State) IsMainParamGroupUsed() bool {
	if s.parameterGroup == nil {
		return false
	}

	for _, memberCluster := range s.memberClusters {
		if memberCluster.CacheParameterGroup == nil {
			continue
		}
		if ptr.Deref(memberCluster.CacheParameterGroup.CacheParameterGroupName, "") == ptr.Deref(s.parameterGroup.CacheParameterGroupName, "") {
			return true
		}
	}

	return false
}

func (s *State) IsTempParamGroupUsed() bool {
	if s.tempParameterGroup == nil {
		return false
	}

	for _, memberCluster := range s.memberClusters {
		if memberCluster.CacheParameterGroup == nil {
			continue
		}
		if ptr.Deref(memberCluster.CacheParameterGroup.CacheParameterGroupName, "") == ptr.Deref(s.tempParameterGroup.CacheParameterGroupName, "") {
			return true
		}
	}

	return false
}

func (s *State) GetUpgradeParamGroupName() string {
	paramGroupName := ptr.Deref(s.parameterGroup.CacheParameterGroupName, "")

	if !s.IsMainParamGroupFamilyUpToDate() && s.tempParameterGroup != nil {
		paramGroupName = ptr.Deref(s.tempParameterGroup.CacheParameterGroupName, "")
	}

	return paramGroupName
}

func (s *State) IsShardCountUpToDate() bool {
	if s.elastiCacheReplicationGroup.NodeGroups == nil {
		return false
	}

	desiredShards := s.ObjAsRedisCluster().Spec.Instance.Aws.ShardCount
	currentShards := int32(len(s.elastiCacheReplicationGroup.NodeGroups))

	return desiredShards == currentShards
}

func (s *State) GetShardsForRemoval() []string {
	if s.elastiCacheReplicationGroup.NodeGroups == nil {
		return nil
	}

	desiredShards := s.ObjAsRedisCluster().Spec.Instance.Aws.ShardCount
	currentShards := int32(len(s.elastiCacheReplicationGroup.NodeGroups))

	if desiredShards >= currentShards {
		return nil
	}

	shardsToRemoveCount := int(currentShards - desiredShards)
	nodeGroups := s.elastiCacheReplicationGroup.NodeGroups
	nodeGroupsForRemoval := nodeGroups[len(nodeGroups)-shardsToRemoveCount:]

	return pie.Map(nodeGroupsForRemoval, func(nodeGroupForRemoval elasticacheTypes.NodeGroup) string {
		return ptr.Deref(nodeGroupForRemoval.NodeGroupId, "")
	})
}

func (s *State) IsReplicaCountUpToDate() bool {
	if s.elastiCacheReplicationGroup.NodeGroups == nil {
		return false
	}

	desiredReplicas := s.ObjAsRedisCluster().Spec.Instance.Aws.ReplicasPerShard

	for _, nodeGroup := range s.elastiCacheReplicationGroup.NodeGroups {
		if nodeGroup.NodeGroupMembers == nil {
			return false
		}

		currentReplicas := int32(len(nodeGroup.NodeGroupMembers) - 1) // substracting one for the master node
		if currentReplicas != desiredReplicas {
			return false
		}
	}

	return true
}

func (s *State) GetReplicasForRemoval() []string {
	if s.elastiCacheReplicationGroup.NodeGroups == nil {
		return []string{}
	}

	desiredReplicas := s.ObjAsRedisCluster().Spec.Instance.Aws.ReplicasPerShard
	replicasToRemove := []string{}

	for _, nodeGroup := range s.elastiCacheReplicationGroup.NodeGroups {
		nodeGroupMembers := nodeGroup.NodeGroupMembers
		if nodeGroupMembers == nil {
			continue
		}

		currentReplicas := int32(len(nodeGroupMembers) - 1) // substracting one for the master node
		if desiredReplicas >= currentReplicas {
			continue
		}

		replicasToRemoveCount := int(currentReplicas - desiredReplicas)
		nodeGroupMembersToRemove := nodeGroupMembers[len(nodeGroupMembers)-replicasToRemoveCount:]
		nodeGroupMemberCacheClusterIds := pie.Map(nodeGroupMembersToRemove, func(nodeGroupMember elasticacheTypes.NodeGroupMember) string {
			return ptr.Deref(nodeGroupMember.CacheClusterId, "")
		})
		replicasToRemove = append(replicasToRemove, nodeGroupMemberCacheClusterIds...)
	}

	return replicasToRemove
}
