package cloudresources

import (
	"context"
	"fmt"

	cloudwatchlogstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	awsconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/config"
	awsutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/util"
	cloudresourcestypes "github.com/kyma-project/cloud-manager/pkg/skr/cloudresources/types"
	awscloudresourcesclient "github.com/kyma-project/cloud-manager/pkg/skr/provider/aws/cloudresources/client"
)

type StateFactory interface {
	NewState(ctx context.Context, crState cloudresourcestypes.State) (*State, error)
}

func NewStateFactory(
	awsClientProvider awsclient.SkrClientProvider[awscloudresourcesclient.Client],
) StateFactory {
	return &stateFactory{
		awsClientProvider: awsClientProvider,
	}
}

type stateFactory struct {
	awsClientProvider awsclient.SkrClientProvider[awscloudresourcesclient.Client]
}

func (f *stateFactory) NewState(ctx context.Context, crState cloudresourcestypes.State) (*State, error) {
	roleName := awsutil.RoleArnDefault(crState.Scope().Spec.Scope.Aws.AccountId)
	c, err := f.awsClientProvider(
		ctx,
		crState.Scope().Spec.Scope.Aws.AccountId,
		crState.Scope().Spec.Region,
		awsconfig.AwsConfig.Default.AccessKeyId,
		awsconfig.AwsConfig.Default.SecretAccessKey,
		roleName,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create aws client: %w", err)
	}
	return &State{
		State:     crState,
		awsClient: c,
	}, nil
}

type State struct {
	cloudresourcestypes.State

	awsClient awscloudresourcesclient.Client

	logGroup *cloudwatchlogstypes.LogGroup
}

// RetentionDays returns the number of days to retain the log events in the WAF log group.
// From aws-sdk-go-v2/service/cloudwatchlogs/api_op_PutRetentionPolicy.go:
// Possible values are: 1, 3, 5, 7, 14, 30, 60, 90, 120, 150, 180, 365, 400, 545,
// 731, 1096, 1827, 2192, 2557, 2922, 3288, and 3653.
func (s *State) RetentionDays() int32 {
	return 7
}
