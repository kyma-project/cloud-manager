package awsnfsvolumebackup

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	backuptypes "github.com/aws/aws-sdk-go-v2/service/backup/types"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsClient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	backupclient "github.com/kyma-project/cloud-manager/pkg/skr/awsnfsvolumebackup/client"
	commonScope "github.com/kyma-project/cloud-manager/pkg/skr/common/scope"
	ctrl "sigs.k8s.io/controller-runtime"
)

type State struct {
	commonScope.State
	awsClientProvider awsClient.SkrClientProvider[backupclient.Client]
	env               abstractions.Environment
	roleName          string

	awsClient         backupclient.Client
	skrAwsNfsVolume   *cloudresourcesv1beta1.AwsNfsVolume
	kcpAwsNfsInstance *cloudcontrolv1beta1.NfsInstance
	vault             *backuptypes.BackupVaultListMember
	backupJob         *backup.DescribeBackupJobOutput
	recoveryPoint     *backup.DescribeRecoveryPointOutput
	//vaultTags         map[string]string
}

func newStateFactory(
	composedStateFactory composed.StateFactory,
	commonScopeStateFactory commonScope.StateFactory,
	awsClientProvider awsClient.SkrClientProvider[backupclient.Client],
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
	commonScopeStateFactory commonScope.StateFactory
	awsClientProvider       awsClient.SkrClientProvider[backupclient.Client]
	env                     abstractions.Environment
}

func (f *stateFactory) NewState(req ctrl.Request) *State {
	return &State{
		State: f.commonScopeStateFactory.NewState(
			f.composedStateFactory.NewState(req.NamespacedName, &cloudresourcesv1beta1.AwsNfsVolumeBackup{}),
		),
		awsClientProvider: f.awsClientProvider,
		env:               f.env,
	}
}

func (s *State) ObjAsAwsNfsVolumeBackup() *cloudresourcesv1beta1.AwsNfsVolumeBackup {
	return s.Obj().(*cloudresourcesv1beta1.AwsNfsVolumeBackup)
}

func (s *State) GetVaultName() string {
	return fmt.Sprintf("cm-%s", s.Scope().Name)
}
func (s *State) GetFileSystemArn() string {
	if s.kcpAwsNfsInstance == nil {
		return ""
	}
	arn := fmt.Sprintf("arn:aws:elasticfilesystem:%s:%s:file-system/%s",
		s.Scope().Spec.Region, s.Scope().Spec.Scope.Aws.AccountId, s.kcpAwsNfsInstance.Status.Id)
	return arn
}

func (s *State) GetRecoveryPointArn() string {
	id := s.ObjAsAwsNfsVolumeBackup().Status.Id
	if id == "" {
		return ""
	}
	arn := fmt.Sprintf("arn:aws:backup:%s:%s:recovery-point:%s",
		s.Scope().Spec.Region, s.Scope().Spec.Scope.Aws.AccountId, id)
	return arn
}

func (s *State) GetBackupName() string {
	return fmt.Sprintf("%s/%s",
		s.Obj().GetNamespace(), s.Obj().GetName())
}

func (s *State) GetTags() map[string]string {
	return map[string]string{
		"Name":                     s.GetBackupName(),
		common.TagCloudManagerName: s.Name().String(),
		common.TagScope:            s.Scope().Name,
		common.TagShoot:            s.Scope().Spec.ShootName,
	}
}

func (f *stateFactory) test() {
	var _ composed.State = State{}
	var _ commonScope.State = State{}
	var _ composed.State = (*State)(nil)
	var _ commonScope.State = (*State)(nil)
}
