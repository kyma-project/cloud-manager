package iprange

import (
	"context"
	"fmt"

	iprangetypes "github.com/kyma-project/cloud-manager/pkg/kcp/iprange/types"
	alicloudconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/config"
	alicloudiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/iprange/client"
)

type State struct {
	iprangetypes.State

	client alicloudiprangeclient.Client

	vpcId     string
	vSwitchId string
	vSwitch   *alicloudiprangeclient.VSwitchInfo
}

type StateFactory interface {
	NewState(ctx context.Context, ipRangeState iprangetypes.State) (*State, error)
}

func NewStateFactory(clientProvider alicloudiprangeclient.ClientProvider) StateFactory {
	return &stateFactory{
		clientProvider: clientProvider,
	}
}

type stateFactory struct {
	clientProvider alicloudiprangeclient.ClientProvider
}

func (f *stateFactory) NewState(ctx context.Context, ipRangeState iprangetypes.State) (*State, error) {
	accessKeyId := alicloudconfig.AlicloudConfig.AccessKeyId
	accessKeySecret := alicloudconfig.AlicloudConfig.AccessKeySecret
	region := ipRangeState.Scope().Spec.Region

	c, err := f.clientProvider(ctx, region, accessKeyId, accessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("error creating alicloud iprange client: %w", err)
	}

	return &State{
		State:  ipRangeState,
		client: c,
	}, nil
}

func (s *State) VSwitchName() string {
	return fmt.Sprintf("cm-%s", s.ObjAsIpRange().Name)
}
