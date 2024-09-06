package redisinstance

import (
	"context"
	"fmt"

	elasticacheTypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	secretsmanager "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
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

	c, err := f.skrProvider(
		ctx,
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
