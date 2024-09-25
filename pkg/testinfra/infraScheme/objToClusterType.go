package infraScheme

import (
	"github.com/kyma-project/cloud-manager/pkg/testinfra/infraTypes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ObjToClusterType(obj client.Object) infraTypes.ClusterType {
	for ct, sch := range SchemeMap {
		_, _, err := sch.ObjectKinds(obj)
		if err == nil {
			return ct
		}
	}
	return ""
}
