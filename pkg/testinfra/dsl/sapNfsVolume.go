package dsl

import (
	"context"
	"errors"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateSapNfsVolume(ctx context.Context, clnt client.Client, obj *cloudresourcesv1beta1.SapNfsVolume, opts ...ObjAction) error {
	if obj == nil {
		obj = &cloudresourcesv1beta1.SapNfsVolume{}
	}
	NewObjActions(opts...).
		Append(
			WithNamespace(DefaultSkrNamespace),
		).
		ApplyOnObject(obj)

	if obj.Name == "" {
		return errors.New("the SKR SapNfsVolume must have name set")
	}

	err := clnt.Create(ctx, obj)
	return err
}

func WithSapNfsVolumeCapacity(capacityGb int) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.SapNfsVolume); ok {
				if x.Spec.CapacityGb == 0 {
					x.Spec.CapacityGb = capacityGb
				}
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithSapNfsVolumeCapacity", obj))
		},
	}
}

func WithSapNfsVolumePvLabels(pvLabels map[string]string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.SapNfsVolume); ok {
				if x.Spec.PersistentVolume == nil {
					x.Spec.PersistentVolume = &cloudresourcesv1beta1.NameLabelsAnnotationsSpec{}
				}
				x.Spec.PersistentVolume.Labels = pvLabels
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithSapNfsVolumePvLabels", obj))
		},
	}
}

func WithSapNfsVolumePvAnnotations(pvAnnotations map[string]string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.SapNfsVolume); ok {
				if x.Spec.PersistentVolume == nil {
					x.Spec.PersistentVolume = &cloudresourcesv1beta1.NameLabelsAnnotationsSpec{}
				}
				x.Spec.PersistentVolume.Annotations = pvAnnotations
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithSapNfsVolumePvLabels", obj))
		},
	}
}

func WithSapNfsVolumePvcLabels(pvcLabels map[string]string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.SapNfsVolume); ok {
				if x.Spec.PersistentVolumeClaim == nil {
					x.Spec.PersistentVolumeClaim = &cloudresourcesv1beta1.NameLabelsAnnotationsSpec{}
				}
				x.Spec.PersistentVolumeClaim.Labels = pvcLabels
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithSapNfsVolumePvcLabels", obj))
		},
	}
}

func WithSapNfsVolumePvcAnnotations(pvcAnnotations map[string]string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.SapNfsVolume); ok {
				if x.Spec.PersistentVolumeClaim == nil {
					x.Spec.PersistentVolumeClaim = &cloudresourcesv1beta1.NameLabelsAnnotationsSpec{}
				}
				x.Spec.PersistentVolumeClaim.Annotations = pvcAnnotations
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithSapNfsVolumePvcLabels", obj))
		},
	}
}

func HavingSapNfsVolumeStatusId() ObjAssertion {
	return func(obj client.Object) error {
		x, ok := obj.(*cloudresourcesv1beta1.SapNfsVolume)
		if !ok {
			return fmt.Errorf("the object %T is not SKR SapNfsVolume", obj)
		}
		if x.Status.Id == "" {
			return errors.New("the SKR SapNfsVolume ID not set")
		}
		return nil
	}
}

func HavingSapNfsVolumeStatusState(state string) ObjAssertion {
	return func(obj client.Object) error {
		x, ok := obj.(*cloudresourcesv1beta1.SapNfsVolume)
		if !ok {
			return fmt.Errorf("the object %T is not SKR SapNfsVolume", obj)
		}
		if x.Status.State != state {
			return fmt.Errorf("the SKR SapNfsVolume State does not match. expected: %s, got: %s", state, x.Status.State)
		}
		return nil
	}
}
