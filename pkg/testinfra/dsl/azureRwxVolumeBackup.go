package dsl

import (
	"context"
	"errors"

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
