package dsl

import (
	"context"
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GivenPvExists(ctx context.Context, clnt client.Client, obj *corev1.PersistentVolume, opts ...ObjAction) error {
	if obj == nil {
		obj = &corev1.PersistentVolume{}
	}
	NewObjActions(opts...).
		Append(
			WithNamespace(DefaultSkrNamespace),
		).
		ApplyOnObject(obj)

	if obj.Name == "" {
		return errors.New("the PersistentVolume must have name set")
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

func GivenPvcExists(ctx context.Context, clnt client.Client, obj *corev1.PersistentVolumeClaim, opts ...ObjAction) error {
	if obj == nil {
		obj = &corev1.PersistentVolumeClaim{}
	}
	NewObjActions(opts...).
		Append(
			WithNamespace(DefaultSkrNamespace),
		).
		ApplyOnObject(obj)

	if obj.Name == "" {
		return errors.New("the PersistentVolumeClaim must have name set")
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

func WithPvCapacity(capacity string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*corev1.PersistentVolume); ok {
				x.Spec.Capacity = corev1.ResourceList{
					"storage": resource.MustParse(capacity),
				}
				return
			}
			if x, ok := obj.(*corev1.PersistentVolumeClaim); ok {
				x.Spec.Resources = corev1.VolumeResourceRequirements{
					Requests: corev1.ResourceList{
						"storage": resource.MustParse(capacity),
					},
				}
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithPVCapacity", obj))
		},
	}
}
func WithPvAccessMode(mode corev1.PersistentVolumeAccessMode) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*corev1.PersistentVolume); ok {
				x.Spec.AccessModes = []corev1.PersistentVolumeAccessMode{mode}
				return
			}
			if x, ok := obj.(*corev1.PersistentVolumeClaim); ok {
				x.Spec.AccessModes = []corev1.PersistentVolumeAccessMode{mode}
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithPVAccessMode", obj))
		},
	}
}
func WithPvCsiSource(src *corev1.CSIPersistentVolumeSource) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*corev1.PersistentVolume); ok {
				x.Spec.CSI = src
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithPVSource", obj))
		},
	}
}

func WithPVName(pvName string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*corev1.PersistentVolumeClaim); ok {
				x.Spec.VolumeName = pvName
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithPVName", obj))
		},
	}
}

func HavePvcPhase(expected corev1.PersistentVolumeClaimPhase) ObjAssertion {
	return func(obj client.Object) error {
		if x, ok := obj.(*corev1.PersistentVolumeClaim); ok {
			actual := x.Status.Phase

			if len(actual) == 0 && actual != expected {
				return fmt.Errorf(
					"expected object %T %s/%s to have phase: %s, but found %s",
					obj, obj.GetNamespace(), obj.GetName(), expected, actual,
				)
			}

		}
		return nil
	}
}
