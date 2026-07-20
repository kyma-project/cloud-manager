package nfsinstance

import (
	"context"
	"fmt"

	nfsinstancetypes "github.com/kyma-project/cloud-manager/pkg/kcp/nfsinstance/types"
	alicloudconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/config"
	alicloudnfsinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/nfsinstance/client"
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

	c, err := f.clientProvider(ctx, region, accessKeyId, accessKeySecret)
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
