package dsl

import (
	"context"
	"errors"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpnfsbackupclientv1 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsbackup/client/v1"
	"google.golang.org/api/file/v1"
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

func CreateGcpFileBackupDirectly(ctx context.Context, backupClient gcpnfsbackupclientv1.FileBackupClient, project, location string, backup *file.Backup) error {
	_, err := backupClient.CreateFileBackup(ctx, project, location, backup.Name, backup)
	return err
}

func ListGcpFileBackups(ctx context.Context, backupClient gcpnfsbackupclientv1.FileBackupClient, project, scopeName string) ([]*file.Backup, error) {
	filter := gcpclient.GetSkrBackupsFilter(scopeName)
	return backupClient.ListFilesBackups(ctx, project, filter)
}

func CreateGcpNfsVolumeBackup(ctx context.Context, clnt client.Client, obj *cloudresourcesv1beta1.GcpNfsVolumeBackup, opts ...ObjAction) error {
	if obj == nil {
		obj = &cloudresourcesv1beta1.GcpNfsVolumeBackup{}
	}
	NewObjActions(WithNamespace(DefaultSkrNamespace),
		WithGcpNfsVolumeBackupValues()).
		Append(
			opts...,
		).ApplyOnObject(obj)

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

func AssertGcpNfsVolumeBackupHasState(state string) ObjAssertion {
	return func(obj client.Object) error {
		x, ok := obj.(*cloudresourcesv1beta1.GcpNfsVolumeBackup)
		if !ok {
			return fmt.Errorf("the object %T is not GcpNfsVolumeBackup", obj)
		}
		if x.Status.State == "" {
			return errors.New("the GcpNfsVolumeBackup state not set")
		}
		if x.Status.State != cloudresourcesv1beta1.GcpNfsBackupState(state) {
			return fmt.Errorf("the GcpNfsVolumeBackup state is %s, expected %s", x.Status.State, state)
		}
		return nil
	}
}

func WithGcpNfsVolumeBackupLocation(location string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.GcpNfsVolumeBackup); ok {
				x.Spec.Location = location
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithGcpNfsVolumeBackupLocation", obj))
		},
	}
}

func WithGcpNfsVolumeBackupAccessibleFrom(accessibleFrom []string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.GcpNfsVolumeBackup); ok {
				x.Spec.AccessibleFrom = accessibleFrom
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithGcpNfsVolumeBackupAccessibleFrom", obj))
		},
	}
}

func HavingGcpNfsVolumeBackupAccessibleFromStatus(accessibleFrom string) ObjAssertion {
	return func(obj client.Object) error {
		x, ok := obj.(*cloudresourcesv1beta1.GcpNfsVolumeBackup)
		if !ok {
			return fmt.Errorf("the object %T is not GcpNfsVolumeBackup", obj)
		}
		if x.Status.AccessibleFrom != accessibleFrom {
			return fmt.Errorf("the GcpNfsVolumeBackup AccessibleFrom status is %s, expected %s", x.Status.AccessibleFrom, accessibleFrom)
		}
		return nil
	}
}
