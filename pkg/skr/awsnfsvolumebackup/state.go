package awsnfsvolumebackup

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/backup"
	backuptypes "github.com/aws/aws-sdk-go-v2/service/backup/types"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	awsconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/config"
	awsutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/util"
	awsnfsvolumebackupclient "github.com/kyma-project/cloud-manager/pkg/skr/awsnfsvolumebackup/client"
	commonscope "github.com/kyma-project/cloud-manager/pkg/skr/common/scope"
	ctrl "sigs.k8s.io/controller-runtime"
)

type State struct {
	commonscope.State
	awsClientProvider awsclient.SkrClientProvider[awsnfsvolumebackupclient.Client]
	env               abstractions.Environment
	roleName          string

	awsClient         awsnfsvolumebackupclient.Client
	skrAwsNfsVolume   *cloudresourcesv1beta1.AwsNfsVolume
	kcpAwsNfsInstance *cloudcontrolv1beta1.NfsInstance
	vault             *backuptypes.BackupVaultListMember
	backupJob         *backup.DescribeBackupJobOutput
	recoveryPoint     *backup.DescribeRecoveryPointOutput
	destAwsClient     awsnfsvolumebackupclient.Client
	destVault         *backuptypes.BackupVaultListMember
	copyJob           *backup.DescribeCopyJobOutput
	destRecoveryPoint *backup.DescribeRecoveryPointOutput
	//vaultTags         map[string]string
}

func newStateFactory(
	composedStateFactory composed.StateFactory,
	commonScopeStateFactory commonscope.StateFactory,
	awsClientProvider awsclient.SkrClientProvider[awsnfsvolumebackupclient.Client],
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
	awsClientProvider       awsclient.SkrClientProvider[awsnfsvolumebackupclient.Client]
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
	if s.Scope() == nil {
		return ""
	}
	return fmt.Sprintf("cm-%s", s.Scope().Name)
}
func (s *State) GetFileSystemArn() string {
	if s.kcpAwsNfsInstance == nil || s.Scope() == nil {
		return ""
	}
	arn := awsutil.EfsArn(s.Scope().Spec.Region, s.Scope().Spec.Scope.Aws.AccountId, s.kcpAwsNfsInstance.Status.Id)
	return arn
}

func (s *State) GetRecoveryPointArn() string {
	id := s.ObjAsAwsNfsVolumeBackup().Status.Id
	if len(id) == 0 || s.Scope() == nil {
		return ""
	}
	arn := awsutil.BackupRecoveryPointArn(s.Scope().Spec.Region, s.Scope().Spec.Scope.Aws.AccountId, id)
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

func (s *State) isTimeForCapacityUpdate() bool {
	lastUpdate := s.ObjAsAwsNfsVolumeBackup().Status.LastCapacityUpdate
	configInterval := awsconfig.AwsConfig.EfsCapacityCheckInterval
	capacityUpdateDue := lastUpdate == nil || lastUpdate.Time.IsZero() || time.Since(lastUpdate.Time) > configInterval
	return capacityUpdateDue
}

func (s *State) requiresRemoteBackup() bool {
	location := s.ObjAsAwsNfsVolumeBackup().Spec.Location
	return len(location) > 0 && s.Scope() != nil && s.Scope().Spec.Region != location
}

func (s *State) GetDestinationBackupVaultArn() string {
	if len(s.ObjAsAwsNfsVolumeBackup().Spec.Location) == 0 || s.Scope() == nil {
		return ""
	}
	arn := awsutil.BackupVaultArn(s.ObjAsAwsNfsVolumeBackup().Spec.Location, s.Scope().Spec.Scope.Aws.AccountId, s.GetVaultName())
	return arn
}

func (s *State) GetDestinationRecoveryPointArn() string {
	id := s.ObjAsAwsNfsVolumeBackup().Status.RemoteId
	if len(s.ObjAsAwsNfsVolumeBackup().Spec.Location) == 0 || s.Scope() == nil || len(id) == 0 {
		return ""
	}

	arn := awsutil.BackupRecoveryPointArn(s.ObjAsAwsNfsVolumeBackup().Spec.Location, s.Scope().Spec.Scope.Aws.AccountId, id)
	return arn
}

func stopAndRequeueForCapacity() error {
	return composed.StopWithRequeueDelay(awsconfig.AwsConfig.EfsCapacityCheckInterval)
}

func StopAndRequeueForCapacityAction() composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		return stopAndRequeueForCapacity(), nil
	}
}
