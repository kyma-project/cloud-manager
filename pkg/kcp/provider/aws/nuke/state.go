package nuke

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/backup/types"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	nuketypes "github.com/kyma-project/cloud-manager/pkg/kcp/nuke/types"
	awsClient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	awsnukeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/nuke/client"
	"k8s.io/utils/ptr"
)

type StateFactory interface {
	NewState(ctx context.Context, nukeState nuketypes.State) (focal.State, error)
}

func NewStateFactory(
	awsClientProvider awsClient.SkrClientProvider[awsnukeclient.NukeNfsBackupClient],
	env abstractions.Environment) StateFactory {
	return stateFactory{
		awsClientProvider: awsClientProvider,
		env:               env,
	}
}

type stateFactory struct {
	awsClientProvider awsClient.SkrClientProvider[awsnukeclient.NukeNfsBackupClient]
	env               abstractions.Environment
}

func (f stateFactory) NewState(ctx context.Context, nukeState nuketypes.State) (focal.State, error) {
	return &State{
		State:             nukeState,
		awsClientProvider: f.awsClientProvider,
		env:               f.env,
	}, nil
}

type State struct {
	nuketypes.State
	ProviderResources []*nuketypes.ProviderResourceKindState

	awsClientProvider awsClient.SkrClientProvider[awsnukeclient.NukeNfsBackupClient]
	env               abstractions.Environment
	awsClient         awsnukeclient.NukeNfsBackupClient
}

type AwsBackup struct {
	*types.RecoveryPointByBackupVault
}

func (b AwsBackup) GetId() string {
	return ptr.Deref(b.RecoveryPointArn, "")
}

func (b AwsBackup) GetObject() interface{} {
	return b
}

type ProviderNukeStatus struct {
	v1beta1.NukeStatus
}

func (s *State) GetVaultName() string {
	return fmt.Sprintf("cm-%s", s.Scope().Name)
}

func (s *State) GetAccountId() string {
	return s.Scope().Spec.Scope.Aws.AccountId
}
