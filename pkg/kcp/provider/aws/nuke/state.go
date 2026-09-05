package nuke

import (
	"context"
	"fmt"

	acmtypes "github.com/aws/aws-sdk-go-v2/service/acm/types"
	"github.com/aws/aws-sdk-go-v2/service/backup/types"
	wafv2types "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	nuketypes "github.com/kyma-project/cloud-manager/pkg/kcp/nuke/types"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	awsnukeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/nuke/client"
	"k8s.io/utils/ptr"
)

type StateFactory interface {
	NewState(ctx context.Context, nukeState nuketypes.State) (focal.State, error)
}

func NewStateFactory(
	awsClientProvider awsclient.SkrClientProvider[awsnukeclient.NukeClient],
	env abstractions.Environment) StateFactory {
	return stateFactory{
		awsClientProvider: awsClientProvider,
		env:               env,
	}
}

type stateFactory struct {
	awsClientProvider awsclient.SkrClientProvider[awsnukeclient.NukeClient]
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

	vault             *types.BackupVaultListMember
	awsClientProvider awsclient.SkrClientProvider[awsnukeclient.NukeClient]
	env               abstractions.Environment
	awsClient         awsnukeclient.NukeClient
}

type AwsBackup struct {
	*types.RecoveryPointByBackupVault
}

func (b AwsBackup) GetId() string {
	return ptr.Deref(b.RecoveryPointArn, "")
}

func (b AwsBackup) GetObject() any {
	return b
}

type AwsWebAclResource struct {
	Summary wafv2types.WebACLSummary
	Detail  *wafv2types.WebACL // loaded when needed for lock token
}

func (r AwsWebAclResource) GetId() string {
	return ptr.Deref(r.Summary.ARN, "")
}

func (r AwsWebAclResource) GetObject() any {
	return r
}

type AwsCertificateResource struct {
	Summary acmtypes.CertificateSummary
}

func (r AwsCertificateResource) GetId() string {
	return ptr.Deref(r.Summary.CertificateArn, "")
}

func (r AwsCertificateResource) GetObject() any {
	return r
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

// Helper function to check if a WAFv2 tag exists with the expected value
func hasTag(tags []wafv2types.Tag, key, value string) bool {
	for _, tag := range tags {
		if ptr.Deref(tag.Key, "") == key && ptr.Deref(tag.Value, "") == value {
			return true
		}
	}
	return false
}

// Helper function to check if an ACM tag exists with the expected value
func hasCertificateTag(tags []acmtypes.Tag, key, value string) bool {
	for _, tag := range tags {
		if ptr.Deref(tag.Key, "") == key && ptr.Deref(tag.Value, "") == value {
			return true
		}
	}
	return false
}
