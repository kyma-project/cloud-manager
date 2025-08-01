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

func (s *State) GetBackupRoleArn() string {
	arn := fmt.Sprintf("arn:aws:iam::%s:role/%s",
		s.Scope().Spec.Scope.Aws.AccountId, awsconfig.AwsConfig.BackupRoleName)
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
	backup := s.ObjAsAwsNfsVolumeBackup()

	lastUpdate := backup.Status.LastCapacityUpdate
	configInterval := awsconfig.AwsConfig.EfsCapacityCheckInterval
	capacityUpdateDue := lastUpdate == nil || lastUpdate.Time.IsZero() || time.Since(lastUpdate.Time) > configInterval
	fmt.Println("Capacity Update Due:", lastUpdate, ", ", configInterval, ", ", capacityUpdateDue)
	return capacityUpdateDue
}

func stopAndRequeueForCapacity() error {
	return composed.StopWithRequeueDelay(awsconfig.AwsConfig.EfsCapacityCheckInterval)
}

func StopAndRequeueForCapacityAction() composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		return stopAndRequeueForCapacity(), nil
	}
}
