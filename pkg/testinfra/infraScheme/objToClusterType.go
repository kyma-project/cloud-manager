package infraScheme

import (
	"github.com/kyma-project/cloud-manager/pkg/testinfra/infraTypes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ObjToClusterType(obj client.Object) infraTypes.ClusterType {
	if obj == nil {
		return ""
	}
	// objects used as unstructured are not in the scheme and must be handled by GVK
	if obj.GetObjectKind().GroupVersionKind().String() == "infrastructuremanager.kyma-project.io/v1, Kind=GardenerCluster" {
		return infraTypes.ClusterTypeKcp
	}
	if obj.GetObjectKind().GroupVersionKind().String() == "operator.kyma-project.io/v1beta2, Kind=Kyma" {
		return infraTypes.ClusterTypeKcp
	}
	for ct, sch := range SchemeMap {
		_, _, err := sch.ObjectKinds(obj)
		if err == nil {
			return ct
		}
	}
	return ""
}
