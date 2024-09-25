package v1beta1

type ScopeRef struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="Scope is immutable."
	// +kubebuilder:validation:XValidation:rule=(self != ""), message="Scope is required."
	Name string `json:"name"`
}
