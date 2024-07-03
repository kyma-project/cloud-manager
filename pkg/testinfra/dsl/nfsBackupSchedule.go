package dsl

import (
	"context"
	"errors"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

func CreateNfsBackupSchedule(ctx context.Context, clnt client.Client, obj *cloudresourcesv1beta1.NfsBackupSchedule, opts ...ObjAction) error {
	if obj == nil {
		obj = &cloudresourcesv1beta1.NfsBackupSchedule{}
	}
	NewObjActions(opts...).
		Append(
			WithNamespace(DefaultSkrNamespace),
		).
		ApplyOnObject(obj)

	if obj.Name == "" {
		return errors.New("the SKR NfsBackupSchedule must have name set")
	}
	if obj.Spec.NfsVolumeRef.Name == "" {
		return errors.New("the SKR NfsBackupSchedule must have spec.NfsVolumeRef.name set")
	}
	if obj.Spec.NfsVolumeRef.Namespace == "" {
		obj.Spec.NfsVolumeRef.Namespace = DefaultSkrNamespace
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
			if x, ok := obj.(*cloudresourcesv1beta1.NfsBackupSchedule); ok {
				x.Spec.Schedule = schedule
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithSchedule", obj))

		},
	}
}
func WithLocation(location string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.NfsBackupSchedule); ok {
				x.Spec.Location = location
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithLocation", obj))

		},
	}
}
func WithNfsVolumeRef(volumeName string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.NfsBackupSchedule); ok {
				x.Spec.NfsVolumeRef.Name = volumeName
				if x.Spec.NfsVolumeRef.Namespace == "" {
					x.Spec.NfsVolumeRef.Namespace = DefaultSkrNamespace
				}
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithNfsVolumeRef", obj))
		},
	}
}
func WithStartTime(start time.Time) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.NfsBackupSchedule); ok {
				x.Spec.StartTime = &metav1.Time{Time: start}
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithStartTime", obj))

		},
	}
}

func WithRetentionDays(days int) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.NfsBackupSchedule); ok {
				x.Spec.MaxRetentionDays = days
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithRetentionDays", obj))
		},
	}
}

func HaveNextRunTimes(expectedTimes []time.Time) ObjAssertion {
	return func(obj client.Object) error {
		if x, ok := obj.(*cloudresourcesv1beta1.NfsBackupSchedule); ok {
			if len(x.Status.NextRunTimes) != len(expectedTimes) {
				return fmt.Errorf(
					"expected object %T %s/%s to have %d runtimes set, but found %d",
					obj, obj.GetNamespace(), obj.GetName(),
					len(expectedTimes), len(x.Status.NextRunTimes),
				)
			}
			for i, t := range expectedTimes {
				actual := x.Status.NextRunTimes[i]
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
			if x, ok := obj.(*cloudresourcesv1beta1.NfsBackupSchedule); ok {
				t := runTime.UTC().Format(time.RFC3339)
				if len(x.Status.NextRunTimes) == 0 {
					x.Status.NextRunTimes = []string{t}
				} else {
					x.Status.NextRunTimes[0] = t
				}
			}
		},
	}
}

func WithNfsBackups(refs ...*corev1.ObjectReference) ObjStatusAction {
	return &objStatusAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.NfsBackupSchedule); ok {

				for _, ref := range refs {
					x.Status.Backups = append(x.Status.Backups, *ref)
				}
			}
		},
	}
}
