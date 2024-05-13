package dsl

import (
	"context"
	"errors"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func WithGcpNfsVolume(name string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.GcpNfsVolumeBackup); ok {
				x.Spec.Source.Volume.Name = name
				if x.Spec.Source.Volume.Namespace == "" {
					x.Spec.Source.Volume.Namespace = DefaultSkrNamespace
				}
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithGcpNfsVolume", obj))
		},
	}
}

func CreateGcpNfsVolumeBackup(ctx context.Context, clnt client.Client, obj *cloudresourcesv1beta1.GcpNfsVolumeBackup, opts ...ObjAction) error {
	if obj == nil {
		obj = &cloudresourcesv1beta1.GcpNfsVolumeBackup{}
	}
	NewObjActions(opts...).
		Append(
			WithNamespace(DefaultSkrNamespace),
			WithGcpNfsVolumeBackupValues(),
		).
		ApplyOnObject(obj)

	if obj.Name == "" {
		return errors.New("the SKR GcpNfsVolumeBackup must have name set")
	}
	if obj.Spec.Source.Volume.Name == "" {
		return errors.New("the SKR GcpNfsVolumeBackup must have spec.source.volume.name set")
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

func WithGcpNfsVolumeBackupValues() ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.GcpNfsVolumeBackup); ok {
				x.Spec.Location = "us-west1"
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithGcpNfsVolumeBackupValues", obj))
		},
	}
}

func AssertGcpNfsVolumeBackupHasStatus() ObjAssertion {
	return func(obj client.Object) error {
		x, ok := obj.(*cloudresourcesv1beta1.GcpNfsVolumeBackup)
		if !ok {
			return fmt.Errorf("the object %T is not GcpNfsVolumeBackup", obj)
		}
		if x.Status.State == "" {
			return errors.New("the GcpNfsVolumeBackup state not set")
		}
		return nil
	}
}

func WithGcpNfsVolumeBackupState(state string) ObjStatusAction {
	return &objStatusAction{
		f: func(obj client.Object) {
			x := obj.(*cloudresourcesv1beta1.GcpNfsVolumeBackup)
			x.Status.State = cloudresourcesv1beta1.GcpNfsBackupState(state)
		},
	}
}
