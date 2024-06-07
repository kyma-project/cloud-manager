package gcpnfsvolume

import (
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
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
	result := map[string]string{
		cloudresourcesv1beta1.LabelNfsVolName: gcpVol.Name,
		cloudresourcesv1beta1.LabelNfsVolNS:   gcpVol.Namespace,
	}
	if gcpVol.Spec.PersistentVolume != nil {
		for k, v := range gcpVol.Spec.PersistentVolume.Labels {
			result[k] = v
		}
	}
	result[cloudresourcesv1beta1.LabelCloudManaged] = "true"
	return result
}

func getVolumeAnnotations(gcpVol *cloudresourcesv1beta1.GcpNfsVolume) map[string]string {
	if gcpVol.Spec.PersistentVolume == nil {
		return nil
	}
	result := map[string]string{}
	for k, v := range gcpVol.Spec.PersistentVolume.Annotations {
		result[k] = v
	}
	return result
}

func getVolumeClaimLabels(gcpVol *cloudresourcesv1beta1.GcpNfsVolume) map[string]string {
	result := map[string]string{
		cloudresourcesv1beta1.LabelNfsVolName: gcpVol.Name,
		cloudresourcesv1beta1.LabelNfsVolNS:   gcpVol.Namespace,
	}
	if gcpVol.Spec.PersistentVolumeClaim != nil {
		for k, v := range gcpVol.Spec.PersistentVolumeClaim.Labels {
			result[k] = v
		}
	}
	result[cloudresourcesv1beta1.LabelCloudManaged] = "true"
	return result
}

func getVolumeClaimAnnotations(gcpVol *cloudresourcesv1beta1.GcpNfsVolume) map[string]string {
	if gcpVol.Spec.PersistentVolumeClaim == nil {
		return nil
	}
	result := map[string]string{}
	for k, v := range gcpVol.Spec.PersistentVolumeClaim.Annotations {
		result[k] = v
	}
	return result
}
