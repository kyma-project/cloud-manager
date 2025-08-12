package dsl

import (
	"context"
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	azure "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/vpcpeering/client"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateAzureVpcPeering(ctx context.Context, clnt client.Client, obj *cloudresourcesv1beta1.AzureVpcPeering, opts ...ObjAction) error {
	NewObjActions(opts...).
		Append(
			WithNamespace(DefaultSkrNamespace),
		).
		ApplyOnObject(obj)

	err := clnt.Create(ctx, obj)
	return err
}

func WithAzureRemotePeeringName(remotePeeringName string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			x := obj.(*cloudresourcesv1beta1.AzureVpcPeering)
			x.Spec.RemotePeeringName = remotePeeringName
		},
	}
}

func WithAzureRemoteVnet(remoteVnet string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			x := obj.(*cloudresourcesv1beta1.AzureVpcPeering)
			x.Spec.RemoteVnet = remoteVnet
		},
	}
}

func AssertAzureVpcPeeringHasId() ObjAssertion {
	return func(obj client.Object) error {
		x, ok := obj.(*cloudresourcesv1beta1.AzureVpcPeering)
		if !ok {
			return fmt.Errorf("the object %T is not  AzureVpcPeering", obj)
		}
		if x.Status.Id == "" {
			return errors.New("the AzureVpcPeering ID not set")
		}
		return nil
	}
}

type GenericAssertion[T any] func(obj T) error

type GenericAssertions[T any] []GenericAssertion[T]

func (x GenericAssertions[T]) AssertObj(obj T) error {
	for _, f := range x {
		if err := f(obj); err != nil {
			return err
		}
	}
	return nil
}

func NewGenericAssertions[T any](items []GenericAssertion[T]) GenericAssertions[T] {
	return append(GenericAssertions[T]{}, items...)
}

func LoadAzureVpcPeeringAndCheck(ctx context.Context, clnt azure.Client, resourceGroupName, virtualNetworkName, virtualNetworkPeeringName string, asserts ...GenericAssertion[armnetwork.VirtualNetworkPeering]) error {

	peering, err := clnt.GetPeering(ctx, resourceGroupName, virtualNetworkName, virtualNetworkPeeringName)

	if err != nil {
		return err
	}

	return NewGenericAssertions(asserts).AssertObj(*peering)
}

func HasLocalAddressPrefix(addressPrefix string) GenericAssertion[armnetwork.VirtualNetworkPeering] {
	return func(obj armnetwork.VirtualNetworkPeering) error {
		actual := ptr.Deref(obj.Properties.LocalAddressSpace.AddressPrefixes[0], "")
		if actual != addressPrefix {
			return fmt.Errorf("expected peering address prefix %s to be %s", actual, addressPrefix)
		}
		return nil
	}
}

func HasRemoteAddressPrefix(addressPrefix string) GenericAssertion[armnetwork.VirtualNetworkPeering] {
	return func(obj armnetwork.VirtualNetworkPeering) error {
		actual := ptr.Deref(obj.Properties.RemoteAddressSpace.AddressPrefixes[0], "")
		if actual != addressPrefix {
			return fmt.Errorf("expected peering address prefix %s to be %s", actual, addressPrefix)
		}
		return nil
	}
}

func HasPeeringSyncLevel(level armnetwork.VirtualNetworkPeeringLevel) GenericAssertion[armnetwork.VirtualNetworkPeering] {
	return func(obj armnetwork.VirtualNetworkPeering) error {
		actual := ptr.Deref(obj.Properties.PeeringSyncLevel, "")
		if actual != level {
			return fmt.Errorf("expected peering sync level %s to be %s", actual, level)
		}
		return nil
	}
}
