package awsnfsvolumebackup

import (
	"github.com/aws/aws-sdk-go-v2/service/backup"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
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

	awsClient         backupclient.Client
	skrAwsNfsVolume   *cloudresourcesv1beta1.AwsNfsVolume
	kcpAwsNfsInstance *cloudcontrolv1beta1.NfsInstance
	vault             *backup.DescribeBackupVaultOutput
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
