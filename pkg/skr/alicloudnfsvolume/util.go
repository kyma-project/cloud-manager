package alicloudnfsvolume

import (
	"maps"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func getVolumeName(aliVol *cloudresourcesv1beta1.AlicloudNfsVolume) string {
	if aliVol.Spec.PersistentVolume != nil &&
		len(aliVol.Spec.PersistentVolume.Name) > 0 {
		return aliVol.Spec.PersistentVolume.Name
	}

	return aliVol.Status.Id
}

func getVolumeLabels(aliVol *cloudresourcesv1beta1.AlicloudNfsVolume) map[string]string {
	labelsBuilder := util.NewLabelBuilder()

	if aliVol.Spec.PersistentVolume != nil {
		for labelName, labelValue := range aliVol.Spec.PersistentVolume.Labels {
			labelsBuilder.WithCustomLabel(labelName, labelValue)
		}
	}

	labelsBuilder.WithCustomLabel(cloudresourcesv1beta1.LabelNfsVolName, aliVol.Name)
	labelsBuilder.WithCustomLabel(cloudresourcesv1beta1.LabelNfsVolNS, aliVol.Namespace)
	labelsBuilder.WithCustomLabel(cloudresourcesv1beta1.LabelCloudManaged, "true")
	labelsBuilder.WithCloudManagerDefaults()
	pvLabels := labelsBuilder.Build()

	return pvLabels
}

func getVolumeAnnotations(aliVol *cloudresourcesv1beta1.AlicloudNfsVolume) map[string]string {
	if aliVol.Spec.PersistentVolume == nil {
		return nil
	}
	result := map[string]string{}
	maps.Copy(result, aliVol.Spec.PersistentVolume.Annotations)
	return result
}

func getVolumeClaimName(aliVol *cloudresourcesv1beta1.AlicloudNfsVolume) string {
	if aliVol.Spec.PersistentVolumeClaim != nil &&
		len(aliVol.Spec.PersistentVolumeClaim.Name) > 0 {
		return aliVol.Spec.PersistentVolumeClaim.Name
	}

	return aliVol.Name
}

func getVolumeClaimLabels(aliVol *cloudresourcesv1beta1.AlicloudNfsVolume) map[string]string {
	labelsBuilder := util.NewLabelBuilder()

	if aliVol.Spec.PersistentVolumeClaim != nil {
		for labelName, labelValue := range aliVol.Spec.PersistentVolumeClaim.Labels {
			labelsBuilder.WithCustomLabel(labelName, labelValue)
		}
	}

	labelsBuilder.WithCustomLabel(cloudresourcesv1beta1.LabelNfsVolName, aliVol.Name)
	labelsBuilder.WithCustomLabel(cloudresourcesv1beta1.LabelNfsVolNS, aliVol.Namespace)
	labelsBuilder.WithCustomLabel(cloudresourcesv1beta1.LabelCloudManaged, "true")
	labelsBuilder.WithCloudManagerDefaults()
	labelsBuilder.WithCustomLabel(cloudresourcesv1beta1.LabelStorageCapacity, aliVol.Spec.Capacity.String())

	pvcLabels := labelsBuilder.Build()
	return pvcLabels
}

func getVolumeClaimAnnotations(aliVol *cloudresourcesv1beta1.AlicloudNfsVolume) map[string]string {
	if aliVol.Spec.PersistentVolumeClaim == nil {
		return nil
	}
	result := map[string]string{}
	maps.Copy(result, aliVol.Spec.PersistentVolumeClaim.Annotations)
	return result
}
