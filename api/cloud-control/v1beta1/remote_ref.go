package v1beta1

import "fmt"

type RemoteRef struct {
	// +kubebuilder:validation:Required
	Namespace string `json:"namespace"`

	// +kubebuilder:validation:Required
	Name string `json:"name"`
}

func (rr RemoteRef) String() string {
	return fmt.Sprintf("%s--%s", rr.Namespace, rr.Name)
}
