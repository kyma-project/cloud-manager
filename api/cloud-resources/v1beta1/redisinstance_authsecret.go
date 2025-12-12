package v1beta1

type RedisAuthSecretSpec struct {
	// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`
	// +kubebuilder:validation:MaxLength=253
	// +kubebuilder:validation:XValidation:rule="self == '' || oldSelf == '' || self == oldSelf",message="name is immutable"
	Name string `json:"name,omitempty"`
	// +kubebuilder:validation:MaxProperties=64
	Labels map[string]string `json:"labels,omitempty"`
	// +kubebuilder:validation:MaxProperties=64
	Annotations map[string]string `json:"annotations,omitempty"`
	ExtraData   map[string]string `json:"extraData,omitempty"`
}
