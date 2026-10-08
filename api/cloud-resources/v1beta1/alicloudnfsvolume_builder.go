package v1beta1

import "k8s.io/apimachinery/pkg/api/resource"

func NewAlicloudNfsVolumeBuilder() *AlicloudNfsVolumeBuilder {
	return &AlicloudNfsVolumeBuilder{
		AlicloudNfsVolume: AlicloudNfsVolume{
			Spec: AlicloudNfsVolumeSpec{},
		},
	}
}

// +kubebuilder:object:generate=false

type AlicloudNfsVolumeBuilder struct {
	AlicloudNfsVolume AlicloudNfsVolume
}

func (b *AlicloudNfsVolumeBuilder) Reset() *AlicloudNfsVolumeBuilder {
	b.AlicloudNfsVolume = AlicloudNfsVolume{
		Spec: AlicloudNfsVolumeSpec{},
	}
	return b
}

func (b *AlicloudNfsVolumeBuilder) WithIpRange(ipRangeName string) *AlicloudNfsVolumeBuilder {
	b.AlicloudNfsVolume.Spec.IpRange.Name = ipRangeName
	return b
}

func (b *AlicloudNfsVolumeBuilder) WithCapacity(capacity string) *AlicloudNfsVolumeBuilder {
	b.AlicloudNfsVolume.Spec.Capacity = resource.MustParse(capacity)
	return b
}

func (b *AlicloudNfsVolumeBuilder) WithStorageType(storageType AlicloudNfsStorageType) *AlicloudNfsVolumeBuilder {
	b.AlicloudNfsVolume.Spec.StorageType = storageType
	return b
}

func (b *AlicloudNfsVolumeBuilder) WithPersistentVolume(name string, labels, annotations map[string]string) *AlicloudNfsVolumeBuilder {
	b.AlicloudNfsVolume.Spec.PersistentVolume = &AlicloudNfsVolumePvSpec{
		Name:        name,
		Labels:      labels,
		Annotations: annotations,
	}
	return b
}

func (b *AlicloudNfsVolumeBuilder) WithPersistentVolumeClaim(name string, labels, annotations map[string]string) *AlicloudNfsVolumeBuilder {
	b.AlicloudNfsVolume.Spec.PersistentVolumeClaim = &AlicloudNfsVolumePvcSpec{
		Name:        name,
		Labels:      labels,
		Annotations: annotations,
	}
	return b
}
