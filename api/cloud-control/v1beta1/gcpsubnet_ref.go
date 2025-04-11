package v1beta1

type GcpSubnetRef struct {
	// +kubebuilder:validation:Required
	Name string `json:"name"`
}

const (
	GcpSubnetField = ".spec.subnet"
)
