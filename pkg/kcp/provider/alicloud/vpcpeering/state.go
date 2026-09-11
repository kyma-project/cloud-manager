package vpcpeering

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	alicloudconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/config"
	alicloudvpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/vpcpeering/client"
	vpcpeeringtypes "github.com/kyma-project/cloud-manager/pkg/kcp/vpcpeering/types"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

	vpcPeering        *alicloudvpcpeeringclient.VpcPeerInfo
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
	localRegion := scope.Spec.Region
	localAccountId := scope.Spec.Scope.Alicloud.AccountId

	// The local VPC ID is stored in VpcNetwork.Status.Identifiers.Vpc, not in the Scope.
	// The VpcNetwork object has the same name as the scope (the kymaName).
	localVpcId, err := loadLocalVpcId(ctx, vpcPeeringState, scope.Name, scope.Namespace)
	if err != nil {
		logger.Error(err, "Error loading local VPC ID from VpcNetwork")
		// Continue with empty localVpcId; createVpcPeerConnection will detect and requeue
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

func loadLocalVpcId(ctx context.Context, vpcPeeringState vpcpeeringtypes.State, scopeName, scopeNamespace string) (string, error) {
	vpcNetwork := &cloudcontrolv1beta1.VpcNetwork{}
	err := vpcPeeringState.Cluster().K8sClient().Get(ctx, types.NamespacedName{
		Name:      scopeName,
		Namespace: scopeNamespace,
	}, vpcNetwork)
	if client.IgnoreNotFound(err) != nil {
		return "", fmt.Errorf("error loading VpcNetwork %s/%s: %w", scopeNamespace, scopeName, err)
	}
	if err != nil {
		return "", nil // VpcNetwork not found yet, will be retried
	}
	return vpcNetwork.Status.Identifiers.Vpc, nil
}
