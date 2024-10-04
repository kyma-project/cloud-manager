package redisinstance

import (
	"context"
	"fmt"

	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elasticacheTypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	secretsmanager "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	awsconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/config"
	client "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/redisinstance/client"
	"github.com/kyma-project/cloud-manager/pkg/kcp/redisinstance/types"
	"k8s.io/utils/ptr"
)

type State struct {
	types.State
	awsClient client.ElastiCacheClient

	subnetGroup                 *elasticacheTypes.CacheSubnetGroup
	parameterGroup              *elasticacheTypes.CacheParameterGroup
	elastiCacheReplicationGroup *elasticacheTypes.ReplicationGroup
	authTokenValue              *secretsmanager.GetSecretValueOutput
	userGroup                   *elasticacheTypes.UserGroup
	securityGroup               *ec2Types.SecurityGroup
	securityGroupId             string

	modifyElastiCacheClusterOptions client.ModifyElastiCacheClusterOptions
	updateMask                      []string
}

type StateFactory interface {
	NewState(ctx context.Context, redisInstace types.State) (*State, error)
}

func NewStateFactory(skrProvider awsclient.SkrClientProvider[client.ElastiCacheClient]) StateFactory {
	return &stateFactory{
		skrProvider: skrProvider,
	}
}

type stateFactory struct {
	skrProvider awsclient.SkrClientProvider[client.ElastiCacheClient]
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

func newState(redisInstace types.State, elastiCacheClient client.ElastiCacheClient) *State {
	return &State{
		State:                           redisInstace,
		awsClient:                       elastiCacheClient,
		modifyElastiCacheClusterOptions: client.ModifyElastiCacheClusterOptions{},
		updateMask:                      []string{},
	}
}

func (s *State) ShouldUpdateRedisInstance() bool {
	return len(s.updateMask) > 0
}

func (s *State) GetModifyElastiCacheClusterOptions() client.ModifyElastiCacheClusterOptions {
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

func (s *State) UpdateTransitEncryptionEnabled(transitEncryptionEnabled bool, isMidstep bool) {

	if isMidstep {
		s.modifyElastiCacheClusterOptions.TransitEncryptionMode = ptr.To(elasticacheTypes.TransitEncryptionModePreferred)
		s.updateMask = append(s.updateMask, "transitEncryptionMode")

		if transitEncryptionEnabled {
			s.modifyElastiCacheClusterOptions.TransitEncryptionEnabled = ptr.To(transitEncryptionEnabled)
			s.updateMask = append(s.updateMask, "transitEncryptionEnabled")
		}

		return
	}

	if transitEncryptionEnabled {
		s.modifyElastiCacheClusterOptions.TransitEncryptionMode = ptr.To(elasticacheTypes.TransitEncryptionModeRequired)
		s.updateMask = append(s.updateMask, "transitEncryptionMode")
	}

	s.modifyElastiCacheClusterOptions.TransitEncryptionEnabled = ptr.To(transitEncryptionEnabled)
	s.updateMask = append(s.updateMask, "transitEncryptionEnabled")
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
