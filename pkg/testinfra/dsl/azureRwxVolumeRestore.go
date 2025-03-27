package dsl

import (
	"context"
	"errors"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateAzureRwxVolumeRestore(ctx context.Context, clnt client.Client, obj *cloudresourcesv1beta1.AzureRwxVolumeRestore, opts ...ObjAction) error {
	if obj == nil {
		obj = &cloudresourcesv1beta1.AzureRwxVolumeRestore{}
	}
	NewObjActions(opts...).
		Append(
			WithNamespace(DefaultSkrNamespace),
		).
		ApplyOnObject(obj)

	if obj.Name == "" {
		return errors.New("the SKR AzureRwxVolumeRestore must have name set")
	}
	if obj.Spec.Source.Backup.Name == "" {
		return errors.New("the SKR AzureRwxVolumeRestore must have spec.source.backup.name set")
	}
	if obj.Spec.Source.Backup.Namespace == "" {
		obj.Spec.Source.Backup.Namespace = DefaultSkrNamespace
	}
	if obj.Spec.Destination.Pvc.Name == "" {
		return errors.New("the SKR AzureRwxVolumeRestore must have spec.target.pvc.name set")
	}
	if obj.Spec.Destination.Pvc.Namespace == "" {
		obj.Spec.Destination.Pvc.Namespace = DefaultSkrNamespace
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

func WithSourceBackupRef(name, namespace string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.AzureRwxVolumeRestore); ok {
				x.Spec.Source.Backup.Name = name
				x.Spec.Source.Backup.Namespace = namespace
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithSourceBackup", obj))
		},
	}
}

func WithDestinationPvcRef(name, namespace string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.AzureRwxVolumeRestore); ok {
				x.Spec.Destination.Pvc.Name = name
				x.Spec.Destination.Pvc.Namespace = namespace
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithDestinationPvc", obj))
		},
	}
}

func WithAzureRwxVolumeRestoreState(state string) ObjStatusAction {
	return &objStatusAction{
		f: func(obj client.Object) {
			x := obj.(*cloudresourcesv1beta1.AzureRwxVolumeRestore)
			x.Status.State = state
		},
	}
}
