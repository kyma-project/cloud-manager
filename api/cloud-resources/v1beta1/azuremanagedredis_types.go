package v1beta1

import (
	featuretypes "github.com/kyma-project/cloud-manager/pkg/feature/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// AzureManagedRedisSpec defines the desired state of AzureManagedRedis.
type AzureManagedRedisSpec struct {
	// RedisTier defines the Kyma service tier. The letter encodes the workload class:
	// S = single-node dev, P = HA production, C = clustered (sharded) HA.
	// All tiers map to a single underlying Azure Managed Redis cluster.
	// Scaling within the same family is allowed (e.g. S1→S3, P1→P3, C3→C5).
	// Switching between families is not allowed because it changes the clustering
	// policy or high-availability mode, which Azure does not support after creation.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self[0] == oldSelf[0]), message="redisTier family is immutable: switching between S, P, and C families is not allowed."
	RedisTier AzureManagedRedisTier `json:"redisTier"`

	// AuthSecret customises the name and labels of the generated connection secret.
	// +optional
	AuthSecret *RedisAuthSecretSpec `json:"authSecret,omitempty"`

	// IpRange references the IpRange resource used for private connectivity.
	// If omitted, the SKR controller selects the default IpRange.
	// +optional
	IpRange IpRangeRef `json:"ipRange,omitempty"`
}

// AzureManagedRedisStatus defines the observed state of AzureManagedRedis.
type AzureManagedRedisStatus struct {
	// +optional
	Id string `json:"id,omitempty"`

	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

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
// +kubebuilder:printcolumn:name="Tier",type="string",JSONPath=".spec.redisTier"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"

// AzureManagedRedis is the Schema for the AzureManagedRedis API.
//
// Beta: this resource is available only per request for SAP-internal
// teams and may change in incompatible ways before GA.
type AzureManagedRedis struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AzureManagedRedisSpec   `json:"spec,omitempty"`
	Status AzureManagedRedisStatus `json:"status,omitempty"`
}

func (in *AzureManagedRedis) GetIpRangeRef() IpRangeRef {
	return in.Spec.IpRange
}

func (in *AzureManagedRedis) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

func (in *AzureManagedRedis) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *AzureManagedRedis) SpecificToFeature() featuretypes.FeatureName {
	return featuretypes.FeatureAzureManagedRedis
}

func (in *AzureManagedRedis) SpecificToProviders() []string {
	return []string{"azure"}
}

func (in *AzureManagedRedis) State() string {
	return in.Status.State
}

func (in *AzureManagedRedis) SetState(v string) {
	in.Status.State = v
}

func (in *AzureManagedRedis) CloneForPatchStatus() client.Object {
	result := &AzureManagedRedis{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AzureManagedRedis",
			APIVersion: GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: in.Namespace,
			Name:      in.Name,
		},
		Status: in.Status,
	}
	if result.Status.Conditions == nil {
		result.Status.Conditions = []metav1.Condition{}
	}
	return result
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
