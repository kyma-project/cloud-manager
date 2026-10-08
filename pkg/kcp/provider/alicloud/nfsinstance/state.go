package nfsinstance

import (
	"context"
	"fmt"

	nfsinstancetypes "github.com/kyma-project/cloud-manager/pkg/kcp/nfsinstance/types"
	alicloudconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/config"
	alicloudmetrics "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/metrics"
	alicloudnfsinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/nfsinstance/client"
	alicloudutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/util"
)

type State struct {
	nfsinstancetypes.State

	client alicloudnfsinstanceclient.Client

	fileSystemId    string
	fileSystem      *alicloudnfsinstanceclient.FileSystemInfo
	mountTargets    []alicloudnfsinstanceclient.MountTargetInfo
	accessGroupName string
	accessGroup     *alicloudnfsinstanceclient.AccessGroupInfo
}

type StateFactory interface {
	NewState(ctx context.Context, nfsInstanceState nfsinstancetypes.State) (*State, error)
}

func NewStateFactory(clientProvider alicloudnfsinstanceclient.ClientProvider) StateFactory {
	return &stateFactory{
		clientProvider: clientProvider,
	}
}

type stateFactory struct {
	clientProvider alicloudnfsinstanceclient.ClientProvider
}

func (f *stateFactory) NewState(ctx context.Context, nfsInstanceState nfsinstancetypes.State) (*State, error) {
	accessKeyId := alicloudconfig.AlicloudConfig.AccessKeyId
	accessKeySecret := alicloudconfig.AlicloudConfig.AccessKeySecret
	region := nfsInstanceState.Scope().Spec.Region

	accountId := alicloudmetrics.AccountIdFromScope(nfsInstanceState.Scope())
	if accountId == "" {
		return nil, fmt.Errorf("scope %q for AliCloud NfsInstance has no account id", nfsInstanceState.Scope().Name)
	}
	ctx = alicloudmetrics.AccountIdIntoContext(ctx, accountId)

	c, err := f.clientProvider(ctx, region, accessKeyId, accessKeySecret, alicloudutil.RoleArnDefault(accountId))
	if err != nil {
		return nil, fmt.Errorf("error creating alicloud nfsinstance client: %w", err)
	}

	return &State{
		State:  nfsInstanceState,
		client: c,
	}, nil
}

// AccessGroupName returns the deterministic NAS permission group name for this instance.
func (s *State) AccessGroupName() string {
	return fmt.Sprintf("cm-%s", s.ObjAsNfsInstance().Name)
}
