package v1beta1

type IpRangeRef struct {
	// +kubebuilder:validation:Required
	Name string `json:"name"`
}
