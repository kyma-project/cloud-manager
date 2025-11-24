package gcpnfsvolume

import (
	"fmt"
	"reflect"
	"strings"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/resource"
)

func getVolumeName(gcpVol *cloudresourcesv1beta1.GcpNfsVolume) string {
	if gcpVol.Spec.PersistentVolume != nil &&
		len(gcpVol.Spec.PersistentVolume.Name) > 0 {
		return gcpVol.Spec.PersistentVolume.Name
	}

	return gcpVol.Status.Id
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

// convertBackupUrlToFullPath converts backup URL from format "{location_id}/{backup_id}"
// to GCP full path format "projects/{project}/locations/{location_id}/backups/{backup_id}"
func convertBackupUrlToFullPath(project, backupUrl string) (string, error) {
	parts := strings.Split(backupUrl, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid backup URL format, expected '{location_id}/{backup_id}', got '%s'", backupUrl)
	}

	locationId := parts[0]
	backupId := parts[1]

	if locationId == "" || backupId == "" {
		return "", fmt.Errorf("location_id and backup_id cannot be empty in backup URL '%s'", backupUrl)
	}

	return fmt.Sprintf("projects/%s/locations/%s/backups/%s", project, locationId, backupId), nil
}

// extractBackupLocation extracts just the location_id from the full backup name
// in format: projects/{project_number}/locations/{location_id}/backups/{backup_id}
// Returns the location_id or empty string if format is invalid
func extractBackupLocation(backupUrl string) string {
	parts := strings.Split(backupUrl, "/")
	if len(parts) != 2 {
		return ""
	}

	return parts[0]
}

// extractBackupLocation extracts just the location_id from the full backup name
// in format: projects/{project_number}/locations/{location_id}/backups/{backup_id}
// Returns the location_id or empty string if format is invalid
func extractBackupName(backupUrl string) string {
	parts := strings.Split(backupUrl, "/")
	if len(parts) != 2 {
		return ""
	}

	return parts[1]
}
