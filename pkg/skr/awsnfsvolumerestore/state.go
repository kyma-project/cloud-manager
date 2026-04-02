package awsnfsvolumerestore

import (
	"context"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	awsutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/util"
	restoreclient "github.com/kyma-project/cloud-manager/pkg/skr/awsnfsvolumerestore/client"
	commonscope "github.com/kyma-project/cloud-manager/pkg/skr/common/scope"
	ctrl "sigs.k8s.io/controller-runtime"
)

type State struct {
	commonscope.State
	awsClientProvider awsclient.SkrClientProvider[restoreclient.Client]
	env               abstractions.Environment
	roleName          string

	awsClient             restoreclient.Client
	skrAwsNfsVolume       *cloudresourcesv1beta1.AwsNfsVolume
	skrAwsNfsVolumeBackup *cloudresourcesv1beta1.AwsNfsVolumeBackup
}

func newStateFactory(
	composedStateFactory composed.StateFactory,
	commonScopeStateFactory commonscope.StateFactory,
	awsClientProvider awsclient.SkrClientProvider[restoreclient.Client],
	env abstractions.Environment,
) *stateFactory {
	return &stateFactory{
		composedStateFactory:    composedStateFactory,
		commonScopeStateFactory: commonScopeStateFactory,
		awsClientProvider:       awsClientProvider,
		env:                     env,
	}
}

type stateFactory struct {
	composedStateFactory    composed.StateFactory
	commonScopeStateFactory commonscope.StateFactory
	awsClientProvider       awsclient.SkrClientProvider[restoreclient.Client]
	env                     abstractions.Environment
}

func (f *stateFactory) NewState(ctx context.Context, req ctrl.Request) (*State, error) {
	scopeState, err := f.commonScopeStateFactory.NewState(
		ctx,
		req.NamespacedName,
		f.composedStateFactory.NewState(req.NamespacedName, &cloudresourcesv1beta1.AwsNfsVolumeRestore{}),
	)
	if err != nil {
		return nil, err
	}
	return &State{
		State:             scopeState,
		awsClientProvider: f.awsClientProvider,
		env:               f.env,
	}, nil
}

func (s *State) ObjAsAwsNfsVolumeRestore() *cloudresourcesv1beta1.AwsNfsVolumeRestore {
	return s.Obj().(*cloudresourcesv1beta1.AwsNfsVolumeRestore)
}

func (s *State) GetRecoveryPointArn() string {
	id := s.skrAwsNfsVolumeBackup.Status.Id
	if id == "" {
		return ""
	}
	arnTxt := awsutil.BackupRecoveryPointArn(s.Scope().Spec.Region, s.Scope().Spec.Scope.Aws.AccountId, id)
	return arnTxt
}

func (s *State) GetBackupRoleArn() string {
	arnTxt := awsutil.RoleArnBackup(s.Scope().Spec.Scope.Aws.AccountId)
	return arnTxt
}

func (s *State) GetVaultName() string {
	return fmt.Sprintf("cm-%s", s.Scope().Name)
}
