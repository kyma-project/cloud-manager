package dsl

import (
	"context"
	"errors"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func WithRestoreDestinationVolume(name string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.GcpNfsVolumeRestore); ok {
				x.Spec.Destination.Volume.Name = name
				if x.Spec.Destination.Volume.Namespace == "" {
					x.Spec.Destination.Volume.Namespace = DefaultSkrNamespace
				}
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithRestoreDestinationVolume", obj))
		},
	}
}

func WithRestoreSourceBackup(name string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.GcpNfsVolumeRestore); ok {
				x.Spec.Source.Backup.Name = name
				if x.Spec.Source.Backup.Namespace == "" {
					x.Spec.Source.Backup.Namespace = DefaultSkrNamespace
				}
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithRestoreSourceBackup", obj))
		},
	}
}

func CreateGcpNfsVolumeRestore(ctx context.Context, clnt client.Client, obj *cloudresourcesv1beta1.GcpNfsVolumeRestore, opts ...ObjAction) error {
	if obj == nil {
		obj = &cloudresourcesv1beta1.GcpNfsVolumeRestore{}
	}
	NewObjActions(opts...).
		Append(
			WithNamespace(DefaultSkrNamespace),
		).
		ApplyOnObject(obj)

	if obj.Name == "" {
		return errors.New("the SKR GcpNfsVolumeRestore must have name set")
	}

	if obj.Spec.Destination.Volume.Name == "" {
		return errors.New("the SKR GcpNfsVolumeBackup must have spec.destination.volume.name set")
	}
	if obj.Spec.Destination.Volume.Namespace == "" {
		obj.Spec.Destination.Volume.Namespace = DefaultSkrNamespace
	}

	if obj.Spec.Source.Backup.Name == "" {
		return errors.New("the SKR GcpNfsVolumeBackup must have spec.source.backup.name set")
	}
	if obj.Spec.Source.Backup.Namespace == "" {
		obj.Spec.Source.Backup.Namespace = DefaultSkrNamespace
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

func AssertGcpNfsVolumeRestoreHasState() ObjAssertion {
	return func(obj client.Object) error {
		x, ok := obj.(*cloudresourcesv1beta1.GcpNfsVolumeRestore)
		if !ok {
			return fmt.Errorf("the object %T is not GcpNfsVolumeRestore", obj)
		}
		if x.Status.State == "" {
			return errors.New("the GcpNfsVolumeRestore state not set")
		}
		return nil
	}
}

func WithGcpNfsVolumeRestoreState(state string) ObjStatusAction {
	return &objStatusAction{
		f: func(obj client.Object) {
			x := obj.(*cloudresourcesv1beta1.GcpNfsVolumeRestore)
			x.Status.State = state
		},
	}
}
