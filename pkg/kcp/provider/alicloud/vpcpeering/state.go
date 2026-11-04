package vpcpeering

import (
	"context"
	"fmt"

	vpcpeer "github.com/alibabacloud-go/vpcpeer-20220101/client"
	"github.com/go-logr/logr"
	alicloudconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/config"
	alicloudvpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/vpcpeering/client"
	vpcpeeringtypes "github.com/kyma-project/cloud-manager/pkg/kcp/vpcpeering/types"
)

type State struct {
	vpcpeeringtypes.State

	client       alicloudvpcpeeringclient.Client
	remoteClient alicloudvpcpeeringclient.Client
	provider     alicloudvpcpeeringclient.ClientProvider

	localAccountId  string
	localRegion     string
	localVpcId      string
	remoteRegion    string
	remoteAccountId string

	vpcPeering        *vpcpeer.GetVpcPeerConnectionAttributeResponseBody
	routeTables       []alicloudvpcpeeringclient.RouteTableInfo
	remoteRouteTables []alicloudvpcpeeringclient.RouteTableInfo
}

type StateFactory interface {
	NewState(ctx context.Context, state vpcpeeringtypes.State, logger logr.Logger) (*State, error)
}

func NewStateFactory(clientProvider alicloudvpcpeeringclient.ClientProvider) StateFactory {
	return &stateFactory{clientProvider: clientProvider}
}

type stateFactory struct {
	clientProvider alicloudvpcpeeringclient.ClientProvider
}

func (f *stateFactory) NewState(ctx context.Context, vpcPeeringState vpcpeeringtypes.State, logger logr.Logger) (*State, error) {
	accessKeyId := alicloudconfig.AlicloudConfig.AccessKeyId
	accessKeySecret := alicloudconfig.AlicloudConfig.AccessKeySecret

	scope := vpcPeeringState.Scope()
	if scope.Spec.Scope.Alicloud == nil {
		return nil, fmt.Errorf("scope %s/%s has no AliCloud provider data", scope.Namespace, scope.Name)
	}
	localRegion := scope.Spec.Region
	localAccountId := scope.Spec.Scope.Alicloud.AccountId

	localVpcId := ""
	if ln := vpcPeeringState.LocalNetwork(); ln != nil && ln.Status.Network != nil && ln.Status.Network.Alicloud != nil {
		localVpcId = ln.Status.Network.Alicloud.VpcId
	}

	c, err := f.clientProvider(ctx, localRegion, accessKeyId, accessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("error creating alicloud vpcpeering client: %w", err)
	}

	return &State{
		State:          vpcPeeringState,
		client:         c,
		remoteClient:   c,
		provider:       f.clientProvider,
		localAccountId: localAccountId,
		localRegion:    localRegion,
		localVpcId:     localVpcId,
	}, nil
}
