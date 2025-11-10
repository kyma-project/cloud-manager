package dsl

import (
	"context"
	"errors"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateGcpNfsVolumeBackupDiscovery(ctx context.Context, clnt client.Client, obj *cloudresourcesv1beta1.GcpNfsVolumeBackupDiscovery, opts ...ObjAction) error {
	if obj == nil {
		obj = &cloudresourcesv1beta1.GcpNfsVolumeBackupDiscovery{}
	}
	NewObjActions(opts...).
		Append(
			WithNamespace(DefaultSkrNamespace),
		).
		ApplyOnObject(obj)

	return clnt.Create(ctx, obj)
}

func GivenGcpNfsVolumeBackupDiscoveryExists(ctx context.Context, clnt client.Client, obj *cloudresourcesv1beta1.GcpNfsVolumeBackupDiscovery, opts ...ObjAction) error {
	if obj == nil {
		obj = &cloudresourcesv1beta1.GcpNfsVolumeBackupDiscovery{}
	}
	NewObjActions(opts...).
		Append(
			WithNamespace(DefaultSkrNamespace),
		).
		ApplyOnObject(obj)

	err := clnt.Get(ctx, client.ObjectKeyFromObject(obj), obj)
	if err != nil {
		return CreateGcpNfsVolumeBackupDiscovery(ctx, clnt, obj)
	}
	return nil
}

// WithGcpNfsVolumeBackupDiscoverySpec sets the spec for GcpNfsVolumeBackupDiscovery
func WithGcpNfsVolumeBackupDiscoverySpec(spec cloudresourcesv1beta1.GcpNfsVolumeBackupDiscoverySpec) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			x := obj.(*cloudresourcesv1beta1.GcpNfsVolumeBackupDiscovery)
			x.Spec = spec
		},
	}
}

// AssertGcpNfsVolumeBackupDiscoveryHasState asserts that the discovery has the expected state
func AssertGcpNfsVolumeBackupDiscoveryHasState(state string) ObjAssertion {
	return func(obj client.Object) error {
		x, ok := obj.(*cloudresourcesv1beta1.GcpNfsVolumeBackupDiscovery)
		if !ok {
			return fmt.Errorf("the object %T is not GcpNfsVolumeBackupDiscovery", obj)
		}
		if x.Status.State == "" {
			return errors.New("the GcpNfsVolumeBackupDiscovery state not set")
		}
		if x.Status.State != state {
			return fmt.Errorf("the GcpNfsVolumeBackupDiscovery state is %s, expected %s", x.Status.State, state)
		}
		return nil
	}
}

// AssertGcpNfsVolumeBackupDiscoveryStatusPopulated asserts that the discovery status fields are populated
func AssertGcpNfsVolumeBackupDiscoveryStatusPopulated() ObjAssertion {
	return func(obj client.Object) error {
		x, ok := obj.(*cloudresourcesv1beta1.GcpNfsVolumeBackupDiscovery)
		if !ok {
			return fmt.Errorf("the object %T is not GcpNfsVolumeBackupDiscovery", obj)
		}

		// Check that status fields are populated as expected for a discovery operation
		if x.Status.DiscoverySnapshotTime == nil {
			return fmt.Errorf("expected DiscoverySnapshotTime to be set")
		}

		if x.Status.AvailableBackupsCount == nil {
			return fmt.Errorf("expected AvailableBackupsCount to be set")
		}

		if x.Status.AvailableBackupUris == nil {
			return fmt.Errorf("expected AvailableBackupUris to be set")
		}

		if x.Status.AvailableBackups == nil {
			return fmt.Errorf("expected AvailableBackups to be set")
		}

		return nil
	}
}

// AssertGcpNfsVolumeBackupDiscoveryAvailableBackupsPopulated asserts that AvailableBackups have expected values
func AssertGcpNfsVolumeBackupDiscoveryAvailableBackupsPopulated() ObjAssertion {
	return func(obj client.Object) error {
		x, ok := obj.(*cloudresourcesv1beta1.GcpNfsVolumeBackupDiscovery)
		if !ok {
			return fmt.Errorf("the object %T is not GcpNfsVolumeBackupDiscovery", obj)
		}

		if len(x.Status.AvailableBackups) == 0 {
			return fmt.Errorf("expected AvailableBackups to contain at least one backup")
		}

		for i, backup := range x.Status.AvailableBackups {
			if backup.Uri == "" {
				return fmt.Errorf("AvailableBackups[%d].Uri should not be empty", i)
			}
			if backup.Location == "" {
				return fmt.Errorf("AvailableBackups[%d].Location should not be empty", i)
			}
			if backup.ShootName == "" {
				return fmt.Errorf("AvailableBackups[%d].ShootName should not be empty", i)
			}
			if backup.BackupName == "" {
				return fmt.Errorf("AvailableBackups[%d].BackupName should not be empty", i)
			}
			if backup.BackupNamespace == "" {
				return fmt.Errorf("AvailableBackups[%d].BackupNamespace should not be empty", i)
			}
			if backup.VolumeName == "" {
				return fmt.Errorf("AvailableBackups[%d].VolumeName should not be empty", i)
			}
			if backup.VolumeNamespace == "" {
				return fmt.Errorf("AvailableBackups[%d].VolumeNamespace should not be empty", i)
			}
			if backup.CreationTime == nil {
				return fmt.Errorf("AvailableBackups[%d].CreationTime should not be nil", i)
			}
		}

		return nil
	}
}
