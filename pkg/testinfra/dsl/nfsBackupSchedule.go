package dsl

import (
	"context"
	"errors"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/skr/backupschedule"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

func CreateNfsBackupSchedule(ctx context.Context, clnt client.Client, obj client.Object, opts ...ObjAction) error {
	if obj == nil {
		return errors.New("NfsBackupSchedule object cannot be nil")
	}
	schedule, ok := obj.(backupschedule.BackupSchedule)
	if !ok {
		return errors.New("object should be an instance of backupschedule.BackupSchedule")
	}
	NewObjActions(opts...).
		Append(
			WithNamespace(DefaultSkrNamespace),
		).
		ApplyOnObject(obj)

	if obj.GetName() == "" {
		return errors.New("the SKR GcpNfsBackupSchedule must have name set")
	}

	sourceRef := schedule.GetSourceRef()
	if sourceRef.Name == "" {
		return errors.New("the SKR GcpNfsBackupSchedule must have spec.NfsVolumeRef.name set")
	}
	if sourceRef.Namespace == "" {
		sourceRef.Namespace = DefaultSkrNamespace
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

func WithSchedule(schedule string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(backupschedule.BackupSchedule); ok {
				x.SetSchedule(schedule)
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithSchedule", obj))

		},
	}
}
func WithGcpLocation(location string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.GcpNfsBackupSchedule); ok {
				x.Spec.Location = location
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithGcpLocation", obj))

		},
	}
}
func WithNfsVolumeRef(volumeName string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(backupschedule.BackupSchedule); ok {
				sourceRef := x.GetSourceRef()
				sourceRef.Name = volumeName
				if sourceRef.Namespace == "" {
					sourceRef.Namespace = DefaultSkrNamespace
				}
				x.SetSourceRef(sourceRef)
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithNfsVolumeRef", obj))
		},
	}
}
func WithStartTime(start time.Time) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(backupschedule.BackupSchedule); ok {
				x.SetStartTime(&metav1.Time{Time: start})
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithStartTime", obj))

		},
	}
}

func WithRetentionDays(days int) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(backupschedule.BackupSchedule); ok {
				x.SetMaxRetentionDays(days)
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithRetentionDays", obj))
		},
	}
}

func HaveNextRunTimes(expectedTimes []time.Time) ObjAssertion {
	return func(obj client.Object) error {
		if x, ok := obj.(backupschedule.BackupSchedule); ok {
			nextRunTimes := x.GetNextRunTimes()
			if len(nextRunTimes) != len(expectedTimes) {
				return fmt.Errorf(
					"expected object %T %s/%s to have %d runtimes set, but found %d",
					obj, obj.GetNamespace(), obj.GetName(),
					len(expectedTimes), len(nextRunTimes),
				)
			}
			for i, t := range expectedTimes {
				actual := nextRunTimes[i]
				expected := t.Format(time.RFC3339)
				if actual != expected {
					return fmt.Errorf(
						"expected object %T %s/%s to have %s runtimes set, but found %s",
						obj, obj.GetNamespace(), obj.GetName(), expected, actual,
					)
				}
			}
		}
		return nil
	}
}

func WithNextRunTime(runTime time.Time) ObjStatusAction {
	return &objStatusAction{
		f: func(obj client.Object) {
			if x, ok := obj.(backupschedule.BackupSchedule); ok {
				nextRunTimes := x.GetNextRunTimes()
				t := runTime.UTC().Format(time.RFC3339)
				if len(nextRunTimes) == 0 {
					x.SetNextRunTimes([]string{t})
				} else {
					nextRunTimes[0] = t
				}
			}
		},
	}
}
