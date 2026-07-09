package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AzureManagedRedisSpec defines the desired state of AzureManagedRedis.
type AzureManagedRedisSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="RemoteRef is immutable."
	RemoteRef RemoteRef `json:"remoteRef"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(size(self.name) > 0), message="IpRange name must not be empty."
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="IpRange is immutable."
	IpRange IpRangeRef `json:"ipRange"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(size(self.name) > 0), message="VpcNetwork name must not be empty."
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="VpcNetwork is immutable."
	VpcNetwork corev1.LocalObjectReference `json:"vpcNetwork"`

	// SKU defines the pricing tier.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self.split('_')[0] == oldSelf.split('_')[0]), message="sku family is immutable: changing between Balanced, ComputeOptimized, MemoryOptimized, FlashOptimized, or Enterprise families is not allowed."
	// +kubebuilder:validation:Enum=Balanced_B0;Balanced_B1;Balanced_B3;Balanced_B5;Balanced_B10;Balanced_B20;Balanced_B50;Balanced_B100;Balanced_B150;Balanced_B250;Balanced_B350;Balanced_B500;Balanced_B700;Balanced_B1000;ComputeOptimized_X3;ComputeOptimized_X5;ComputeOptimized_X10;ComputeOptimized_X20;ComputeOptimized_X50;ComputeOptimized_X100;ComputeOptimized_X150;ComputeOptimized_X250;ComputeOptimized_X350;ComputeOptimized_X500;ComputeOptimized_X700;MemoryOptimized_M10;MemoryOptimized_M20;MemoryOptimized_M50;MemoryOptimized_M100;MemoryOptimized_M150;MemoryOptimized_M250;MemoryOptimized_M350;MemoryOptimized_M500;MemoryOptimized_M700;MemoryOptimized_M1000;MemoryOptimized_M1500;MemoryOptimized_M2000;FlashOptimized_A250;FlashOptimized_A500;FlashOptimized_A700;FlashOptimized_A1000;FlashOptimized_A1500;FlashOptimized_A2000;FlashOptimized_A4500;Enterprise_E1;Enterprise_E5;Enterprise_E10;Enterprise_E20;Enterprise_E50;Enterprise_E100;Enterprise_E200;Enterprise_E400;EnterpriseFlash_F300;EnterpriseFlash_F700;EnterpriseFlash_F1500
	SKU string `json:"sku"`

	// HighAvailability enables zone-redundant replica deployment.
	// +optional
	// +kubebuilder:default=true
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="HighAvailability is immutable."
	HighAvailability bool `json:"highAvailability"`

	// ClusteringPolicy defines the Redis clustering mode.
	// +kubebuilder:validation:Enum=EnterpriseCluster;NoCluster;OSSCluster
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="ClusteringPolicy is immutable."
	ClusteringPolicy string `json:"clusteringPolicy"`

	// +kubebuilder:validation:Required
	Scope ScopeRef `json:"scope"`
}

// AzureManagedRedisStatus defines the observed state of AzureManagedRedis.
type AzureManagedRedisStatus struct {
	// +optional
	Id string `json:"id,omitempty"`

	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// PrimaryEndpoint is the connection hostname reported after provisioning.
	// +optional
	PrimaryEndpoint string `json:"primaryEndpoint,omitempty"`

	// Port is the Redis client port (always 10000 for AMR).
	// +optional
	Port int32 `json:"port,omitempty"`

	// AuthString is the access key / password for client authentication.
	// +optional
	AuthString string `json:"authString,omitempty"`

	// List of status conditions
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// +optional
	State string `json:"state,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories={kyma-cloud-manager}
// +kubebuilder:printcolumn:name="VpcNetwork",type="string",JSONPath=".spec.vpcNetwork.name"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"

// AzureManagedRedis is the Schema for the AzureManagedRedis API.
type AzureManagedRedis struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AzureManagedRedisSpec   `json:"spec,omitempty"`
	Status AzureManagedRedisStatus `json:"status,omitempty"`
}

func (in *AzureManagedRedis) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

func (in *AzureManagedRedis) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *AzureManagedRedis) ObservedGeneration() int64 {
	return in.Status.ObservedGeneration
}

func (in *AzureManagedRedis) SetObservedGeneration(v int64) {
	in.Status.ObservedGeneration = v
}

func (in *AzureManagedRedis) GetStatus() any {
	return in.Status
}

func (in *AzureManagedRedis) State() string {
	return in.Status.State
}

func (in *AzureManagedRedis) SetState(v string) {
	in.Status.State = v
}

func (in *AzureManagedRedis) ScopeRef() ScopeRef {
	return in.Spec.Scope
}

func (in *AzureManagedRedis) SetScopeRef(scopeRef ScopeRef) {
	in.Spec.Scope = scopeRef
}

// +kubebuilder:object:root=true

// AzureManagedRedisList contains a list of AzureManagedRedis.
type AzureManagedRedisList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AzureManagedRedis `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AzureManagedRedis{}, &AzureManagedRedisList{})
}
