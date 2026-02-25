package awsnfsvolumerestore

import (
	"fmt"
	"regexp"
	"strings"

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

// efsFileSystemIdRegex validates EFS filesystem ID format: fs-<8-17 hex chars>
// AWS EFS filesystem IDs have format fs-[0-9a-f]{8} or fs-[0-9a-f]{17}
var efsFileSystemIdRegex = regexp.MustCompile(`^fs-[0-9a-f]{8,17}$`)

func (s *State) GetFileSystemId() string {
	if s.skrAwsNfsVolume == nil {
		return ""
	}
	// The AwsNfsVolume.Status.Server contains the EFS filesystem DNS name
	// Format: fs-<id>.efs.<region>.amazonaws.com
	// Extract the filesystem ID (fs-<id>)
	server := s.skrAwsNfsVolume.Status.Server
	if server == "" {
		return ""
	}
	// Extract "fs-xxxxx" from "fs-xxxxx.efs.region.amazonaws.com"
	parts := strings.Split(server, ".")
	if len(parts) > 0 {
		fsId := parts[0]
		// Validate EFS filesystem ID format
		if efsFileSystemIdRegex.MatchString(fsId) {
			return fsId
		}
	}
	return ""
}
