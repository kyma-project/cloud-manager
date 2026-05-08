package v1beta1

import (
	featuretypes "github.com/kyma-project/cloud-manager/pkg/feature/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// AzureManagedRedisInstanceSpec defines the desired state of AzureManagedRedisInstance.
type AzureManagedRedisInstanceSpec struct {
	// SKU defines the pricing tier for Azure Managed Redis.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="sku is immutable."
	SKU AzureManagedRedisSKU `json:"sku"`

	// HighAvailability enables zone-redundant replica deployment.
	// +optional
	// +kubebuilder:default=true
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="highAvailability is immutable."
	HighAvailability bool `json:"highAvailability"`

	// TLSVersion specifies the minimum TLS version for client connections.
	// +optional
	// +kubebuilder:validation:Enum="1.2";"1.3"
	// +kubebuilder:default="1.2"
	TLSVersion string `json:"tlsVersion,omitempty"`

	// ClientProtocol specifies the Redis wire protocol.
	// +optional
	// +kubebuilder:validation:Enum=Plaintext;Encrypted
	// +kubebuilder:default=Encrypted
	ClientProtocol string `json:"clientProtocol,omitempty"`

	// AuthSecret customises the name and labels of the generated connection secret.
	// +optional
	AuthSecret *RedisAuthSecretSpec `json:"authSecret,omitempty"`

	// IpRange references the IpRange resource used for private connectivity.
	// +optional
	IpRange IpRangeRef `json:"ipRange,omitempty"`
}

// AzureManagedRedisInstanceStatus defines the observed state of AzureManagedRedisInstance.
type AzureManagedRedisInstanceStatus struct {
	// +optional
	Id string `json:"id,omitempty"`

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

// AzureManagedRedisInstance is the Schema for the AzureManagedRedisInstances API.
type AzureManagedRedisInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AzureManagedRedisInstanceSpec   `json:"spec,omitempty"`
	Status AzureManagedRedisInstanceStatus `json:"status,omitempty"`
}

func (in *AzureManagedRedisInstance) GetIpRangeRef() IpRangeRef {
	return in.Spec.IpRange
}

func (in *AzureManagedRedisInstance) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

func (in *AzureManagedRedisInstance) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *AzureManagedRedisInstance) SpecificToFeature() featuretypes.FeatureName {
	return featuretypes.FeatureAzureManagedRedis
}

func (in *AzureManagedRedisInstance) SpecificToProviders() []string {
	return []string{"azure"}
}

func (in *AzureManagedRedisInstance) State() string {
	return in.Status.State
}

func (in *AzureManagedRedisInstance) SetState(v string) {
	in.Status.State = v
}

func (in *AzureManagedRedisInstance) CloneForPatchStatus() client.Object {
	result := &AzureManagedRedisInstance{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AzureManagedRedisInstance",
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

// AzureManagedRedisInstanceList contains a list of AzureManagedRedisInstance.
type AzureManagedRedisInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AzureManagedRedisInstance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AzureManagedRedisInstance{}, &AzureManagedRedisInstanceList{})
}
