package dsl

import (
	"errors"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	awsmock "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/mock"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func HavingVpcPeeringStatusId() ObjAssertion {
	return func(obj client.Object) error {
		x, ok := obj.(*cloudcontrolv1beta1.VpcPeering)
		if !ok {
			return fmt.Errorf("the object %T is not KCP VpcPeering", obj)
		}
		if x.Status.Id == "" {
			return errors.New("the KCP VpcPeering .status.id not set")
		}
		return nil
	}
}

func HavingVpcPeeringStatusRemoteId() ObjAssertion {
	return func(obj client.Object) error {
		x, ok := obj.(*cloudcontrolv1beta1.VpcPeering)
		if !ok {
			return fmt.Errorf("the object %T is not KCP VpcPeering", obj)
		}
		if x.Status.RemoteId == "" {
			return errors.New("the KCP VpcPeering .status.remoteId not set")
		}
		return nil
	}
}

func WithRemoteRouteTableUpdateStrategy(strategy cloudcontrolv1beta1.AwsRouteTableUpdateStrategy) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			x := obj.(*cloudcontrolv1beta1.VpcPeering)
			x.Spec.Details.RemoteRouteTableUpdateStrategy = strategy
		},
	}
}

func CheckRoute(config awsmock.RouteTableConfig, vpcId, routeTableId, vpcPeeringConnectionId, destinationCidrBlock string) error {
	route := config.GetRoute(vpcId, routeTableId, vpcPeeringConnectionId, destinationCidrBlock)
	if route == nil {
		return errors.New("route does not exist")
	}
	return nil
}
