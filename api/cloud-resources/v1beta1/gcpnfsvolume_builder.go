package v1beta1

// +kubebuilder:object:generate=false

type GcpNfsVolumeBuilder struct {
	Obj *GcpNfsVolume
}

func NewGcpNfsVolumeBuilder() *GcpNfsVolumeBuilder {
	return (&GcpNfsVolumeBuilder{}).WithObj(&GcpNfsVolume{})
}

func (b *GcpNfsVolumeBuilder) WithObj(obj *GcpNfsVolume) *GcpNfsVolumeBuilder {
	b.Obj = obj
	return b
}

func (b *GcpNfsVolumeBuilder) WithName(name string) *GcpNfsVolumeBuilder {
	b.Obj.Name = name
	return b
}

func (b *GcpNfsVolumeBuilder) WithNamespace(namespace string) *GcpNfsVolumeBuilder {
	b.Obj.Namespace = namespace
	return b
}

func (b *GcpNfsVolumeBuilder) WithIpRange(ipRangeName string) *GcpNfsVolumeBuilder {
	b.Obj.Spec.IpRange.Name = ipRangeName
	return b
}

func (b *GcpNfsVolumeBuilder) WithCapacityGb(capacityGb int) *GcpNfsVolumeBuilder {
	b.Obj.Spec.CapacityGb = capacityGb
	return b
}

func (b *GcpNfsVolumeBuilder) WithTier(tier GcpFileTier) *GcpNfsVolumeBuilder {
	b.Obj.Spec.Tier = tier
	return b
}

func (b *GcpNfsVolumeBuilder) WithFileShareName(fileShareName string) *GcpNfsVolumeBuilder {
	b.Obj.Spec.FileShareName = fileShareName
	return b
}

func (b *GcpNfsVolumeBuilder) WithPvcSpec(name string, labels map[string]string, annotations map[string]string) *GcpNfsVolumeBuilder {
	if b.Obj.Spec.PersistentVolumeClaim == nil {
		b.Obj.Spec.PersistentVolumeClaim = &GcpNfsVolumePvcSpec{}
	}
	b.Obj.Spec.PersistentVolumeClaim.Name = name
	b.Obj.Spec.PersistentVolumeClaim.Labels = labels
	b.Obj.Spec.PersistentVolumeClaim.Annotations = annotations
	return b
}

func (b *GcpNfsVolumeBuilder) WithPvSpec(name string, labels map[string]string, annotations map[string]string) *GcpNfsVolumeBuilder {
	if b.Obj.Spec.PersistentVolume == nil {
		b.Obj.Spec.PersistentVolume = &GcpNfsVolumePvSpec{}
	}
	b.Obj.Spec.PersistentVolume.Name = name
	b.Obj.Spec.PersistentVolume.Labels = labels
	b.Obj.Spec.PersistentVolume.Annotations = annotations
	return b
}

func (b *GcpNfsVolumeBuilder) Build() *GcpNfsVolume {
	return b.Obj
}
