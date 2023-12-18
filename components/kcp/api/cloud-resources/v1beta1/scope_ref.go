package v1beta1

type ScopeRef struct {
	// +kubebuilder:validation:Required
	Name string `json:"name"`
}
