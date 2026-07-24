package dsl

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/filestore/apiv1/filestorepb"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpnfsbackupclientv2 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsbackup/client/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GcpNfsVolumeBackupPath returns the full GCP backup path from scope and backup status.
// Use this to interact with mock GCP APIs in tests.
func GcpNfsVolumeBackupPath(scope *cloudcontrolv1beta1.Scope, backup *cloudresourcesv1beta1.GcpNfsVolumeBackup) string {
	project := scope.Spec.Scope.Gcp.Project
	location := backup.Status.Location
	name := fmt.Sprintf("cm-%.60s", backup.Status.Id)
	return gcpnfsbackupclientv2.GetFileBackupPath(project, location, name)
}

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

// CreateGcpFileBackupDirectly seeds a filestore backup into the mock2 store, bypassing
// source-instance validation. backup.Name is expected to be a bare backup id on input and is
// rewritten in place to the full resource path (projects/{p}/locations/{l}/backups/{id}) so
// callers can assert against the same name the nuke reconciler reports.
func CreateGcpFileBackupDirectly(ctx context.Context, backupClient gcpnfsbackupclientv2.FileBackupClient, project, location string, backup *filestorepb.Backup) error {
	seeder, ok := backupClient.(interface {
		AddFilestoreBackupDirectly(*filestorepb.Backup) error
	})
	if !ok {
		return fmt.Errorf("backup client %T does not support direct backup seeding", backupClient)
	}
	backup.Name = gcpnfsbackupclientv2.GetFileBackupPath(project, location, backup.Name)
	return seeder.AddFilestoreBackupDirectly(backup)
}

func ListGcpFileBackups(ctx context.Context, backupClient gcpnfsbackupclientv2.FileBackupClient, project, scopeName string) ([]*filestorepb.Backup, error) {
	filter := gcpclient.GetSkrBackupsFilter(scopeName)
	parent := gcpnfsbackupclientv2.GetFilestoreParentPath(project, "-")
	iter := backupClient.ListFilestoreBackups(ctx, &filestorepb.ListBackupsRequest{
		Parent: parent,
		Filter: filter,
	})
	var backups []*filestorepb.Backup
	for b, err := range iter.All() {
		if err != nil {
			return nil, err
		}
		backups = append(backups, b)
	}
	return backups, nil
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

func WithGcpNfsVolumeBackupStatusLocation(location string) ObjAction {
	return &objStatusAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.GcpNfsVolumeBackup); ok {
				x.Status.Location = location
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithGcpNfsVolumeBackupStatusLocation", obj))
		},
	}
}

func WithGcpNfsVolumeBackupStatusId(id string) ObjAction {
	return &objStatusAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.GcpNfsVolumeBackup); ok {
				x.Status.Id = id
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithGcpNfsVolumeBackupStatusId", obj))
		},
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
