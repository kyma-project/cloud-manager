package dsl

import (
	"context"
	"errors"
	"fmt"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateAzureRwxVolumeBackup(ctx context.Context, clnt client.Client, obj *cloudresourcesv1beta1.AzureRwxVolumeBackup, opts ...ObjAction) error {
	if obj == nil {
		obj = &cloudresourcesv1beta1.AzureRwxVolumeBackup{}
	}
	NewObjActions(opts...).
		Append(
			WithNamespace(DefaultSkrNamespace),
		).
		ApplyOnObject(obj)

	if obj.Name == "" {
		return errors.New("the SKR AzureRwxVolumeBackup must have name set")
	}
	if obj.Spec.Source.Pvc.Name == "" {
		return errors.New("the SKR AzureRwxVolumeBackup must have spec.source.pvc.name set")
	}
	if obj.Spec.Source.Pvc.Namespace == "" {
		obj.Spec.Source.Pvc.Namespace = DefaultSkrNamespace
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

func GivenAzureRwxVolumeBackupExists(ctx context.Context, clnt client.Client, obj *cloudresourcesv1beta1.AzureRwxVolumeBackup, opts ...ObjAction) error {
	if obj == nil {
		obj = &cloudresourcesv1beta1.AzureRwxVolumeBackup{}
	}
	NewObjActions(opts...).
		Append(
			WithNamespace(DefaultSkrNamespace),
		).
		ApplyOnObject(obj)

	if obj.Name == "" {
		return errors.New("the SKR AzureRwxVolumeBackup must have name set")
	}

	err := clnt.Get(ctx, client.ObjectKeyFromObject(obj), obj)
	if client.IgnoreNotFound(err) != nil {
		return err
	}
	if apierrors.IsNotFound(err) {
		err = clnt.Create(ctx, obj)
	} else {
		err = clnt.Update(ctx, obj)
	}
	return err
}

func WithSourcePvc(name, namespace string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.AzureRwxVolumeBackup); ok {
				x.Spec.Source.Pvc.Name = name
				x.Spec.Source.Pvc.Namespace = namespace
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithSourcePvc", obj))
		},
	}
}

func WithRecoveryPointId(recoveryPointId string) ObjStatusAction {
	return &objStatusAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.AzureRwxVolumeBackup); ok {
				x.Status.RecoveryPointId = recoveryPointId
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithRecoveryPointId", obj))
		},
	}
}

func WithStorageAccountPath(storageAccountPath string) ObjStatusAction {
	return &objStatusAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.AzureRwxVolumeBackup); ok {
				x.Status.StorageAccountPath = storageAccountPath
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithStorageAccountPath", obj))
		},
	}
}
