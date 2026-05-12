package v1beta1

import (
	featuretypes "github.com/kyma-project/cloud-manager/pkg/feature/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// AzureManagedRedisClusterSpec defines the desired state of AzureManagedRedisCluster.
type AzureManagedRedisClusterSpec struct {
	// SKU defines the pricing tier for Azure Managed Redis.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="sku is immutable."
	SKU AzureManagedRedisSKU `json:"sku"`

	// HighAvailability enables zone-redundant replica deployment.
	// +optional
	// +kubebuilder:default=true
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="highAvailability is immutable."
	HighAvailability bool `json:"highAvailability"`

	// AuthSecret customises the name and labels of the generated connection secret.
	// +optional
	AuthSecret *RedisAuthSecretSpec `json:"authSecret,omitempty"`

	// IpRange references the IpRange resource used for private connectivity.
	// +optional
	IpRange IpRangeRef `json:"ipRange,omitempty"`
}

// AzureManagedRedisClusterStatus defines the observed state of AzureManagedRedisCluster.
type AzureManagedRedisClusterStatus struct {
	// +optional
	Id string `json:"id,omitempty"`

	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// PrimaryEndpoint is the connection hostname reported after provisioning.
	// +optional
	PrimaryEndpoint string `json:"primaryEndpoint,omitempty"`

	// Port is the Redis client port.
	// +optional
	Port int32 `json:"port,omitempty"`

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
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"

// AzureManagedRedisCluster is the Schema for the AzureManagedRedisClusters API.
type AzureManagedRedisCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AzureManagedRedisClusterSpec   `json:"spec,omitempty"`
	Status AzureManagedRedisClusterStatus `json:"status,omitempty"`
}

func (in *AzureManagedRedisCluster) GetIpRangeRef() IpRangeRef {
	return in.Spec.IpRange
}

func (in *AzureManagedRedisCluster) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

func (in *AzureManagedRedisCluster) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *AzureManagedRedisCluster) SpecificToFeature() featuretypes.FeatureName {
	return featuretypes.FeatureAzureManagedRedis
}

func (in *AzureManagedRedisCluster) SpecificToProviders() []string {
	return []string{"azure"}
}

func (in *AzureManagedRedisCluster) State() string {
	return in.Status.State
}

func (in *AzureManagedRedisCluster) SetState(v string) {
	in.Status.State = v
}

func (in *AzureManagedRedisCluster) CloneForPatchStatus() client.Object {
	result := &AzureManagedRedisCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AzureManagedRedisCluster",
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

// AzureManagedRedisClusterList contains a list of AzureManagedRedisCluster.
type AzureManagedRedisClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AzureManagedRedisCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AzureManagedRedisCluster{}, &AzureManagedRedisClusterList{})
}
