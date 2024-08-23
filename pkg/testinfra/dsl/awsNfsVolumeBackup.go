package dsl

import (
	"context"
	"errors"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateAwsNfsVolumeBackup(ctx context.Context, clnt client.Client, obj *cloudresourcesv1beta1.AwsNfsVolumeBackup, opts ...ObjAction) error {
	if obj == nil {
		obj = &cloudresourcesv1beta1.AwsNfsVolumeBackup{}
	}
	NewObjActions(opts...).
		Append(
			WithNamespace(DefaultSkrNamespace),
		).
		ApplyOnObject(obj)

	if obj.Name == "" {
		return errors.New("the SKR AwsNfsVolumeBackup must have name set")
	}
	if obj.Spec.Source.Volume.Name == "" {
		return errors.New("the SKR AwsNfsVolumeBackup must have spec.source.volume.name set")
	}
	if obj.Spec.Source.Volume.Namespace == "" {
		obj.Spec.Source.Volume.Namespace = DefaultSkrNamespace
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

func WithAwsNfsVolume(name string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.AwsNfsVolumeBackup); ok {
				x.Spec.Source.Volume.Name = name
				if x.Spec.Source.Volume.Namespace == "" {
					x.Spec.Source.Volume.Namespace = DefaultSkrNamespace
				}
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAwsNfsVolume", obj))
		},
	}
}

func AssertAwsNfsVolumeBackupHasState(state string) ObjAssertion {
	return func(obj client.Object) error {
		x, ok := obj.(*cloudresourcesv1beta1.AwsNfsVolumeBackup)
		if !ok {
			return fmt.Errorf("the object %T is not AwsNfsVolumeBackup", obj)
		}
		if x.Status.State == "" {
			return errors.New("the AwsNfsVolumeBackup state not set")
		}
		if x.Status.State != state {
			return fmt.Errorf("the AwsNfsVolumeBackup state is %s, expected %s", x.Status.State, state)
		}
		return nil
	}
}
