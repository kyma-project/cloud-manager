package dsl

import (
	"context"
	"errors"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func WithNfsVolumeIpRange(ipRangeName string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.AwsNfsVolume); ok {
				if x.Spec.IpRange.Name == "" {
					x.Spec.IpRange.Name = ipRangeName
				}
				return
			}
			if x, ok := obj.(*cloudresourcesv1beta1.GcpNfsVolume); ok {
				if x.Spec.IpRange.Name == "" {
					x.Spec.IpRange.Name = ipRangeName
				}
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithNfsVolumeIpRange", obj))
		},
	}
}

func WithAwsNfsVolumePvName(name string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.AwsNfsVolume); ok {
				if x.Spec.PersistentVolume == nil {
					x.Spec.PersistentVolume = &cloudresourcesv1beta1.AwsNfsVolumePvSpec{}
				}
				x.Spec.PersistentVolume.Name = name
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAwsNfsVolumePvName", obj))
		},
	}
}

func WithAwsNfsVolumePvLabels(labels map[string]string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.AwsNfsVolume); ok {
				if x.Spec.PersistentVolume == nil {
					x.Spec.PersistentVolume = &cloudresourcesv1beta1.AwsNfsVolumePvSpec{}
				}
				if x.Spec.PersistentVolume.Labels == nil {
					x.Spec.PersistentVolume.Labels = map[string]string{}
				}
				for k, v := range labels {
					x.Spec.PersistentVolume.Labels[k] = v
				}
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAwsNfsVolumePvLabels", obj))
		},
	}
}

func WithAwsNfsVolumePvAnnotations(annotations map[string]string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.AwsNfsVolume); ok {
				if x.Spec.PersistentVolume == nil {
					x.Spec.PersistentVolume = &cloudresourcesv1beta1.AwsNfsVolumePvSpec{}
				}
				if x.Spec.PersistentVolume.Annotations == nil {
					x.Spec.PersistentVolume.Annotations = map[string]string{}
				}
				for k, v := range annotations {
					x.Spec.PersistentVolume.Annotations[k] = v
				}
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAwsNfsVolumePvAnnotations", obj))
		},
	}
}

func WithAwsNfsVolumePvcName(name string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.AwsNfsVolume); ok {
				if x.Spec.PersistentVolumeClaim == nil {
					x.Spec.PersistentVolumeClaim = &cloudresourcesv1beta1.AwsNfsVolumePvcSpec{}
				}
				x.Spec.PersistentVolumeClaim.Name = name
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAwsNfsVolumePvcName", obj))
		},
	}
}

func WithAwsNfsVolumePvcLabels(labels map[string]string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.AwsNfsVolume); ok {
				if x.Spec.PersistentVolumeClaim == nil {
					x.Spec.PersistentVolumeClaim = &cloudresourcesv1beta1.AwsNfsVolumePvcSpec{}
				}
				if x.Spec.PersistentVolumeClaim.Labels == nil {
					x.Spec.PersistentVolumeClaim.Labels = map[string]string{}
				}
				for k, v := range labels {
					x.Spec.PersistentVolumeClaim.Labels[k] = v
				}
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAwsNfsVolumePvcLabels", obj))
		},
	}
}

func WithAwsNfsVolumePvcAnnotations(annotations map[string]string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.AwsNfsVolume); ok {
				if x.Spec.PersistentVolumeClaim == nil {
					x.Spec.PersistentVolumeClaim = &cloudresourcesv1beta1.AwsNfsVolumePvcSpec{}
				}
				if x.Spec.PersistentVolumeClaim.Annotations == nil {
					x.Spec.PersistentVolumeClaim.Annotations = map[string]string{}
				}
				for k, v := range annotations {
					x.Spec.PersistentVolumeClaim.Annotations[k] = v
				}
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAwsNfsVolumePvcAnnotations", obj))
		},
	}
}

func WithAwsNfsVolumeCapacity(capacity string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.AwsNfsVolume); ok {
				if x.Spec.Capacity.IsZero() {
					x.Spec.Capacity = resource.MustParse(capacity)
				}
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithNfsVolumeCapacity", obj))
		},
	}
}

func CreateAwsNfsVolume(ctx context.Context, clnt client.Client, obj *cloudresourcesv1beta1.AwsNfsVolume, opts ...ObjAction) error {
	if obj == nil {
		obj = &cloudresourcesv1beta1.AwsNfsVolume{}
	}
	NewObjActions(opts...).
		Append(
			WithNamespace(DefaultSkrNamespace),
		).
		ApplyOnObject(obj)

	if obj.Name == "" {
		return errors.New("the SKR AwsNfsVolume must have name set")
	}

	err := clnt.Create(ctx, obj)
	return err
}

func HavingAwsNfsVolumeStatusId() ObjAssertion {
	return func(obj client.Object) error {
		x, ok := obj.(*cloudresourcesv1beta1.AwsNfsVolume)
		if !ok {
			return fmt.Errorf("the object %T is not SKR AwsNfsVolume", obj)
		}
		if x.Status.Id == "" {
			return errors.New("the SKR AwsNfsVolume ID not set")
		}
		return nil
	}
}

func HavingAwsNfsVolumeStatusState(state string) ObjAssertion {
	return func(obj client.Object) error {
		x, ok := obj.(*cloudresourcesv1beta1.AwsNfsVolume)
		if !ok {
			return fmt.Errorf("the object %T is not SKR AwsNfsVolume", obj)
		}
		if x.Status.State != state {
			return fmt.Errorf("the SKR AwsNfsVolume State does not match. expected: %s, got: %s", state, x.Status.State)
		}
		return nil
	}
}
