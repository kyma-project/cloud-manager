package dsl

import (
	"context"
	"errors"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func WithNfsVolumeIpRange(ipRangeName string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.AwsNfsVolume); ok {
				x.Spec.IpRange.Name = ipRangeName
				if x.Spec.IpRange.Namespace == "" {
					x.Spec.IpRange.Namespace = DefaultSkrNamespace
				}
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithNfsVolumeIpRange", obj))
		},
	}
}

func CreateAwsNfsVolume(ctx context.Context, clnt client.Client, obj *cloudresourcesv1beta1.AwsNfsVolume, opts ...ObjAction) error {
	if obj == nil {
		obj = &cloudresourcesv1beta1.AwsNfsVolume{}
	}
	NewObjActions(opts...).
		Append(
			WithNamespace(DefaultSkrNamespace),
		).
		ApplyOnObject(obj)

	if obj.Name == "" {
		return errors.New("the SKR AwsNfsVolume must have name set")
	}
	if obj.Spec.IpRange.Name == "" {
		return errors.New("the SKR AwsNfsVolume must have spec.iprange.name set")
	}
	if obj.Spec.IpRange.Namespace == "" {
		obj.Spec.IpRange.Namespace = DefaultSkrNamespace
	}

	err := clnt.Create(ctx, obj)
	return err
}
