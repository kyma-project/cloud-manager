package dsl

import (
	"context"
	"errors"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func WithGcpNfsVolumeIpRange(ipRangeName string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.GcpNfsVolume); ok {
				x.Spec.IpRange.Name = ipRangeName
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithNfsVolumeIpRange", obj))
		},
	}
}

func CreateGcpNfsVolume(ctx context.Context, clnt client.Client, obj *cloudresourcesv1beta1.GcpNfsVolume, opts ...ObjAction) error {
	if obj == nil {
		obj = &cloudresourcesv1beta1.GcpNfsVolume{}
	}
	NewObjActions(WithNamespace(DefaultSkrNamespace),
		WithGcpNfsValues()).
		Append(
			opts...,
		).
		ApplyOnObject(obj)

	if obj.Name == "" {
		return errors.New("the SKR GcpNfsVolume must have name set")
	}

	err := clnt.Get(ctx, client.ObjectKeyFromObject(obj), obj)
	if err == nil {
		// already exists
		return nil
	}
	if client.IgnoreNotFound(err) != nil {
		// some error
		return err
	}
	err = clnt.Create(ctx, obj)
	return err
}

func WithGcpNfsValues() ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.GcpNfsVolume); ok {
				x.Spec.Tier = "BASIC_HDD"
				x.Spec.CapacityGb = 1024
				x.Spec.FileShareName = "test01"
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithGcpNfsValues", obj))
		},
	}
}

func WithGcpNfsVolumeStatusLocation(location string) ObjAction {
	return &objStatusAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.GcpNfsVolume); ok {
				x.Status.Location = location
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithGcpNfsValues", obj))
		},
	}
}

func WithGcpNfsVolumeStatusId(id string) ObjAction {
	return &objStatusAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.GcpNfsVolume); ok {
				x.Status.Id = id
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithGcpNfsVolumeStatusId", obj))
		},
	}
}
