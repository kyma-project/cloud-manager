package redisinstance

import (
	"context"
	"strings"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elasticachetypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	awsconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/config"
	awsutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/util"
	"github.com/kyma-project/cloud-manager/pkg/kcp/redisinstance/types"
	"k8s.io/utils/ptr"
)

type State struct {
	types.State
	awsClient awsclient.ElastiCacheClient

	subnetGroup                           *elasticachetypes.CacheSubnetGroup
	parameterGroup                        *elasticachetypes.CacheParameterGroup
	parameterGroupCurrentParams           []elasticachetypes.Parameter
	parameterGroupFamilyDefaultParams     []elasticachetypes.Parameter
	tempParameterGroup                    *elasticachetypes.CacheParameterGroup
	tempParameterGroupFamilyDefaultParams []elasticachetypes.Parameter
	tempParameterGroupCurrentParams       []elasticachetypes.Parameter

	elastiCacheReplicationGroup *elasticachetypes.ReplicationGroup
	memberClusters              []elasticachetypes.CacheCluster
	authTokenValue              *secretsmanager.GetSecretValueOutput
	userGroup                   *elasticachetypes.UserGroup
	securityGroup               *ec2types.SecurityGroup
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
	roleName := awsutil.RoleArnDefault(redisInstace.Scope().Spec.Scope.Aws.AccountId)

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

func (s *State) ShouldUpdateRedisInstance() bool {
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

	desiredParametersMap := GetDesiredParameters(defaultParametersMap, s.ObjAsRedisInstance().Spec.Instance.Aws.Parameters)
	forUpdateParameters := GetMissmatchedParameters(currentParametersMap, desiredParametersMap)

	return len(forUpdateParameters) == 0
}

func (s *State) IsMainParamGroupFamilyUpToDate() bool {
	desiredParamGroupFamily := GetAwsElastiCacheParameterGroupFamily(s.ObjAsRedisInstance().Spec.Instance.Aws.EngineVersion)
	currentParamGroupFamily := ptr.Deref(s.parameterGroup.CacheParameterGroupFamily, "")

	return desiredParamGroupFamily == currentParamGroupFamily
}

func (s *State) AreTempParamGroupParamsUpToDate() bool {
	currentParametersMap := MapParameters(s.tempParameterGroupCurrentParams)
	defaultParametersMap := MapParameters(s.tempParameterGroupFamilyDefaultParams)

	desiredParametersMap := GetDesiredParameters(defaultParametersMap, s.ObjAsRedisInstance().Spec.Instance.Aws.Parameters)
	forUpdateParameters := GetMissmatchedParameters(currentParametersMap, desiredParametersMap)

	return len(forUpdateParameters) == 0
}

func (s *State) IsRedisVersionUpToDate() bool {
	for _, memberCluster := range s.memberClusters {
		if memberCluster.CacheParameterGroup == nil {
			return true
		}

		desiredVersion := s.ObjAsRedisInstance().Spec.Instance.Aws.EngineVersion
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

// GetProvisionedMachineType returns the provisioned machine type from the AWS ElastiCache Replication Group
func (s *State) GetProvisionedMachineType() string {
	if len(s.memberClusters) == 0 {
		return ""
	}
	return ptr.Deref(s.memberClusters[0].CacheNodeType, "")
}

// GetProvisionedReplicaCount returns the provisioned replica count from the AWS ElastiCache Replication Group
func (s *State) GetProvisionedReplicaCount() int32 {
	if s.elastiCacheReplicationGroup == nil || len(s.elastiCacheReplicationGroup.MemberClusters) == 0 {
		return 0
	}

	// Subtract 1 to exclude the primary node, counting only replicas
	memberCount := len(s.elastiCacheReplicationGroup.MemberClusters)
	if memberCount <= 1 {
		return 0
	}
	return int32(memberCount - 1)
}
