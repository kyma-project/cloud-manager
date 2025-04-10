package awsnfsvolumerestore

import (
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	awsconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/config"
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

func (f *stateFactory) NewState(req ctrl.Request) *State {
	return &State{
		State: f.commonScopeStateFactory.NewState(
			f.composedStateFactory.NewState(req.NamespacedName, &cloudresourcesv1beta1.AwsNfsVolumeRestore{}),
		),
		awsClientProvider: f.awsClientProvider,
		env:               f.env,
	}
}

func (s *State) ObjAsAwsNfsVolumeRestore() *cloudresourcesv1beta1.AwsNfsVolumeRestore {
	return s.Obj().(*cloudresourcesv1beta1.AwsNfsVolumeRestore)
}

func (s *State) GetRecoveryPointArn() string {
	id := s.skrAwsNfsVolumeBackup.Status.Id
	if id == "" {
		return ""
	}
	arn := fmt.Sprintf("arn:aws:backup:%s:%s:recovery-point:%s",
		s.Scope().Spec.Region, s.Scope().Spec.Scope.Aws.AccountId, id)
	return arn
}

func (s *State) GetBackupRoleArn() string {
	arn := fmt.Sprintf("arn:aws:iam::%s:role/%s",
		s.Scope().Spec.Scope.Aws.AccountId, awsconfig.AwsConfig.BackupRoleName)
	return arn
}

func (s *State) GetVaultName() string {
	return fmt.Sprintf("cm-%s", s.Scope().Name)
}
