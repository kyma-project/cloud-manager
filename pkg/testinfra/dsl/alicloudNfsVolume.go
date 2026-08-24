package dsl

import (
	"context"
	"errors"
	"fmt"
	"maps"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func WithAlicloudNfsVolumeStorageType(storageType cloudresourcesv1beta1.AlicloudNfsStorageType) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.AlicloudNfsVolume); ok {
				x.Spec.StorageType = storageType
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAlicloudNfsVolumeStorageType", obj))
		},
	}
}

func WithAlicloudNfsVolumeCapacity(capacity string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.AlicloudNfsVolume); ok {
				if x.Spec.Capacity.IsZero() {
					x.Spec.Capacity = resource.MustParse(capacity)
				}
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAlicloudNfsVolumeCapacity", obj))
		},
	}
}

func WithAlicloudNfsVolumePvName(name string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.AlicloudNfsVolume); ok {
				if x.Spec.PersistentVolume == nil {
					x.Spec.PersistentVolume = &cloudresourcesv1beta1.AlicloudNfsVolumePvSpec{}
				}
				x.Spec.PersistentVolume.Name = name
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAlicloudNfsVolumePvName", obj))
		},
	}
}

func WithAlicloudNfsVolumePvLabels(labels map[string]string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.AlicloudNfsVolume); ok {
				if x.Spec.PersistentVolume == nil {
					x.Spec.PersistentVolume = &cloudresourcesv1beta1.AlicloudNfsVolumePvSpec{}
				}
				if x.Spec.PersistentVolume.Labels == nil {
					x.Spec.PersistentVolume.Labels = map[string]string{}
				}
				maps.Copy(x.Spec.PersistentVolume.Labels, labels)
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAlicloudNfsVolumePvLabels", obj))
		},
	}
}

func WithAlicloudNfsVolumePvAnnotations(annotations map[string]string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.AlicloudNfsVolume); ok {
				if x.Spec.PersistentVolume == nil {
					x.Spec.PersistentVolume = &cloudresourcesv1beta1.AlicloudNfsVolumePvSpec{}
				}
				if x.Spec.PersistentVolume.Annotations == nil {
					x.Spec.PersistentVolume.Annotations = map[string]string{}
				}
				maps.Copy(x.Spec.PersistentVolume.Annotations, annotations)
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAlicloudNfsVolumePvAnnotations", obj))
		},
	}
}

func WithAlicloudNfsVolumePvcName(name string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.AlicloudNfsVolume); ok {
				if x.Spec.PersistentVolumeClaim == nil {
					x.Spec.PersistentVolumeClaim = &cloudresourcesv1beta1.AlicloudNfsVolumePvcSpec{}
				}
				x.Spec.PersistentVolumeClaim.Name = name
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAlicloudNfsVolumePvcName", obj))
		},
	}
}

func WithAlicloudNfsVolumePvcLabels(labels map[string]string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.AlicloudNfsVolume); ok {
				if x.Spec.PersistentVolumeClaim == nil {
					x.Spec.PersistentVolumeClaim = &cloudresourcesv1beta1.AlicloudNfsVolumePvcSpec{}
				}
				if x.Spec.PersistentVolumeClaim.Labels == nil {
					x.Spec.PersistentVolumeClaim.Labels = map[string]string{}
				}
				maps.Copy(x.Spec.PersistentVolumeClaim.Labels, labels)
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAlicloudNfsVolumePvcLabels", obj))
		},
	}
}

func WithAlicloudNfsVolumePvcAnnotations(annotations map[string]string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.AlicloudNfsVolume); ok {
				if x.Spec.PersistentVolumeClaim == nil {
					x.Spec.PersistentVolumeClaim = &cloudresourcesv1beta1.AlicloudNfsVolumePvcSpec{}
				}
				if x.Spec.PersistentVolumeClaim.Annotations == nil {
					x.Spec.PersistentVolumeClaim.Annotations = map[string]string{}
				}
				maps.Copy(x.Spec.PersistentVolumeClaim.Annotations, annotations)
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAlicloudNfsVolumePvcAnnotations", obj))
		},
	}
}

func CreateAlicloudNfsVolume(ctx context.Context, clnt client.Client, obj *cloudresourcesv1beta1.AlicloudNfsVolume, opts ...ObjAction) error {
	if obj == nil {
		obj = &cloudresourcesv1beta1.AlicloudNfsVolume{}
	}
	NewObjActions(opts...).
		Append(
			WithNamespace(DefaultSkrNamespace),
		).
		ApplyOnObject(obj)

	if obj.Name == "" {
		return errors.New("the SKR AlicloudNfsVolume must have name set")
	}

	err := clnt.Create(ctx, obj)
	return err
}
