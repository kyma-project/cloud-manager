package v1beta1

import (
	"k8s.io/apimachinery/pkg/types"
)

type IpRangeRef struct {
	// +kubebuilder:validation:Required
	Name string `json:"name"`
	// +kubebuilder:validation:Required
	Namespace string `json:"namespace"`
}

const (
	IpRangeField = ".spec.ipRange"
)

func (in *IpRangeRef) ObjKey() types.NamespacedName {
	return types.NamespacedName{
		Namespace: in.Namespace,
		Name:      in.Name,
	}
}
