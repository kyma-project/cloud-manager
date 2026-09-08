package wafpolicy

import (
	"context"

	wafv2types "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	"github.com/kyma-project/cloud-manager/pkg/skr/provider/aws/wafpolicy/client"
	wafpolicytypes "github.com/kyma-project/cloud-manager/pkg/skr/wafpolicy/types"
)

type State struct {
	wafpolicytypes.State
	awsClientProvider awsclient.SkrClientProvider[client.Client]
	env               abstractions.Environment

	awsClient    client.Client
	roleName     string
	awsWebAcl    *wafv2types.WebACL // Loaded AWS WebACL
	lockToken    string             // Transient lock token from loadWebAcl, not persisted
	updateNeeded bool               // Whether update is needed based on spec vs AWS state
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

func (s *State) AwsClient() client.Client {
	return s.awsClient
}

func (s *State) SetAwsClient(c client.Client) {
	s.awsClient = c
}

func (s *State) RoleName() string {
	return s.roleName
}

func (s *State) SetRoleName(name string) {
	s.roleName = name
}

func (s *State) AwsWebAcl() *wafv2types.WebACL {
	return s.awsWebAcl
}

func (s *State) SetAwsWebAcl(acl *wafv2types.WebACL) {
	s.awsWebAcl = acl
}

func (s *State) LockToken() string {
	return s.lockToken
}

func (s *State) SetLockToken(token string) {
	s.lockToken = token
}

func (s *State) UpdateNeeded() bool {
	return s.updateNeeded
}

func (s *State) SetUpdateNeeded(needed bool) {
	s.updateNeeded = needed
}

func (s *State) Env() abstractions.Environment {
	return s.env
}
