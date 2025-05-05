package v1beta1

import "k8s.io/apimachinery/pkg/types"

type GcpSubnetRef struct {
	// +kubebuilder:validation:Required
	Name string `json:"name"`
}

const (
	GcpSubnetField = ".spec.subnet"
)

func (in *GcpSubnetRef) ObjKey() types.NamespacedName {
	return types.NamespacedName{
		Name: in.Name,
	}
}
