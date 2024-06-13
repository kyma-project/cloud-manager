package gcpnfsvolume

import (
	"reflect"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/resource"
)

func getVolumeName(gcpVol *cloudresourcesv1beta1.GcpNfsVolume) string {
	if gcpVol.Spec.PersistentVolume != nil &&
		len(gcpVol.Spec.PersistentVolume.Name) > 0 {
		return gcpVol.Spec.PersistentVolume.Name
	}

	return gcpVol.Name
}

func getVolumeClaimName(gcpVol *cloudresourcesv1beta1.GcpNfsVolume) string {
	if gcpVol.Spec.PersistentVolumeClaim != nil &&
		len(gcpVol.Spec.PersistentVolumeClaim.Name) > 0 {
		return gcpVol.Spec.PersistentVolumeClaim.Name
	}

	return gcpVol.Name
}

func getVolumeLabels(gcpVol *cloudresourcesv1beta1.GcpNfsVolume) map[string]string {
	labelsBuilder := util.NewLabelBuilder()

	if gcpVol.Spec.PersistentVolume != nil {
		for labelName, labelValue := range gcpVol.Spec.PersistentVolume.Labels {
			labelsBuilder.WithCustomLabel(labelName, labelValue)
		}
	}

	labelsBuilder.WithCustomLabel(cloudresourcesv1beta1.LabelNfsVolName, gcpVol.Name)
	labelsBuilder.WithCustomLabel(cloudresourcesv1beta1.LabelNfsVolNS, gcpVol.Namespace)
	labelsBuilder.WithCustomLabel(cloudresourcesv1beta1.LabelCloudManaged, "true")
	labelsBuilder.WithCloudManagerDefaults()
	pvLabels := labelsBuilder.Build()

	return pvLabels
}

func getVolumeAnnotations(gcpVol *cloudresourcesv1beta1.GcpNfsVolume) map[string]string {
	result := map[string]string{}
	if gcpVol.Spec.PersistentVolume == nil {
		return result
	}
	for k, v := range gcpVol.Spec.PersistentVolume.Annotations {
		result[k] = v
	}
	return result
}

func getVolumeClaimLabels(gcpVol *cloudresourcesv1beta1.GcpNfsVolume) map[string]string {
	labelsBuilder := util.NewLabelBuilder()

	if gcpVol.Spec.PersistentVolumeClaim != nil {
		for labelName, labelValue := range gcpVol.Spec.PersistentVolumeClaim.Labels {
			labelsBuilder.WithCustomLabel(labelName, labelValue)
		}
	}

	labelsBuilder.WithCustomLabel(cloudresourcesv1beta1.LabelNfsVolName, gcpVol.Name)
	labelsBuilder.WithCustomLabel(cloudresourcesv1beta1.LabelNfsVolNS, gcpVol.Namespace)
	labelsBuilder.WithCustomLabel(cloudresourcesv1beta1.LabelCloudManaged, "true")
	labelsBuilder.WithCloudManagerDefaults()
	storage := gcpNfsVolumeCapacityToResourceQuantity(gcpVol)
	labelsBuilder.WithCustomLabel(cloudresourcesv1beta1.LabelStorageCapacity, storage.String())

	pvcLabels := labelsBuilder.Build()
	return pvcLabels
}

func getVolumeClaimAnnotations(gcpVol *cloudresourcesv1beta1.GcpNfsVolume) map[string]string {
	result := map[string]string{}
	if gcpVol.Spec.PersistentVolumeClaim == nil {
		return result
	}
	for k, v := range gcpVol.Spec.PersistentVolumeClaim.Annotations {
		result[k] = v
	}
	return result
}

func gcpNfsVolumeCapacityToResourceQuantity(gcpVol *cloudresourcesv1beta1.GcpNfsVolume) *resource.Quantity {
	return resource.NewQuantity(int64(gcpVol.Spec.CapacityGb)*1024*1024*1024, resource.BinarySI)
}

func areLabelsEqual(first, second map[string]string) bool {
	x := first
	y := second

	if x == nil {
		x = map[string]string{}
	}
	if y == nil {
		y = map[string]string{}
	}

	return reflect.DeepEqual(x, y)
}

func areAnnotationsSuperset(superset, subset map[string]string) bool {
	for key, value := range subset {
		if superset[key] != value {
			return false
		}
	}
	return true
}
