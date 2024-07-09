package v1beta1

import (
	"k8s.io/apimachinery/pkg/types"
)

type IpRangeRef struct {
	// +kubebuilder:validation:Required
	Name string `json:"name"`
}

const (
	IpRangeField = ".spec.ipRange"
)

func (in *IpRangeRef) ObjKey() types.NamespacedName {
	return types.NamespacedName{
		Name: in.Name,
	}
}
