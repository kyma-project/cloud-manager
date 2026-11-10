package wafpolicy

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	wafv2types "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	"github.com/kyma-project/cloud-manager/pkg/skr/provider/aws/wafpolicy/client"
	wafpolicytypes "github.com/kyma-project/cloud-manager/pkg/skr/wafpolicy/types"
)

type State struct {
	wafpolicytypes.State
	awsClientProvider awsclient.SkrClientProvider[client.Client]
	env               abstractions.Environment

	awsClient     client.Client
	awsWebAcl     *wafv2types.WebACL               // Loaded AWS WebACL
	lockToken     string                           // Transient lock token from loadWebAcl, not persisted
	updateNeeded  bool                             // Whether update is needed based on spec vs AWS state
	parsedInput   *wafv2.CreateWebACLInput         // Parsed spec.payload, set by parsePayload action
	loggingConfig *wafv2types.LoggingConfiguration // Loaded by loadLoggingConfig; nil = not configured
}

// Ensure State implements wafpolicytypes.State
var _ wafpolicytypes.State = &State{}

type StateFactory interface {
	NewState(ctx context.Context, wafPolicyState wafpolicytypes.State) (*State, error)
}

func NewStateFactory(
	awsClientProvider awsclient.SkrClientProvider[client.Client],
	env abstractions.Environment,
) StateFactory {
	return &stateFactory{
		awsClientProvider: awsClientProvider,
		env:               env,
	}
}

type stateFactory struct {
	awsClientProvider awsclient.SkrClientProvider[client.Client]
	env               abstractions.Environment
}

func (f *stateFactory) NewState(ctx context.Context, wafPolicyState wafpolicytypes.State) (*State, error) {
	return &State{
		State:             wafPolicyState,
		awsClientProvider: f.awsClientProvider,
		env:               f.env,
	}, nil
}

func (s *State) Env() abstractions.Environment {
	return s.env
}

func (s *State) convertTags() []wafv2types.Tag {
	return []wafv2types.Tag{
		{
			Key:   aws.String(common.TagCloudManagerName),
			Value: aws.String(s.Name().String()),
		},
		{
			Key:   aws.String(common.TagScope),
			Value: aws.String(s.Scope().Name),
		},
		{
			Key:   aws.String(common.TagShoot),
			Value: aws.String(s.Scope().Spec.ShootName),
		},
	}
}
