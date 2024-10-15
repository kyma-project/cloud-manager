package dsl

import (
	"context"
	"errors"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateAwsNfsVolumeRestore(ctx context.Context, clnt client.Client, obj *cloudresourcesv1beta1.AwsNfsVolumeRestore, opts ...ObjAction) error {
	if obj == nil {
		obj = &cloudresourcesv1beta1.AwsNfsVolumeRestore{}
	}
	NewObjActions(opts...).
		Append(
			WithNamespace(DefaultSkrNamespace),
		).
		ApplyOnObject(obj)

	if obj.Name == "" {
		return errors.New("the SKR AwsNfsVolumeRestore must have name set")
	}
	if obj.Spec.Source.Backup.Name == "" {
		return errors.New("the SKR AwsNfsVolumeRestore must have spec.source.backup.name set")
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

func WithAwsNfsVolumeBackup(name string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.AwsNfsVolumeRestore); ok {
				x.Spec.Source.Backup.Name = name
				if x.Spec.Source.Backup.Namespace == "" {
					x.Spec.Source.Backup.Namespace = DefaultSkrNamespace
				}
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAwsNfsVolumeBackup", obj))
		},
	}
}

func AssertAwsNfsVolumeRestoreHasState(state string) ObjAssertion {
	return func(obj client.Object) error {
		x, ok := obj.(*cloudresourcesv1beta1.AwsNfsVolumeRestore)
		if !ok {
			return fmt.Errorf("the object %T is not AwsNfsVolumeRestore", obj)
		}
		if x.Status.State == "" {
			return errors.New("the AwsNfsVolumeRestore state not set")
		}
		if x.Status.State != state {
			return fmt.Errorf("the AwsNfsVolumeRestore state is %s, expected %s", x.Status.State, state)
		}
		return nil
	}
}
