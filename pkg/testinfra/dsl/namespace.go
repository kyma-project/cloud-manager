package dsl

import (
	"github.com/kyma-project/cloud-manager/pkg/testinfra"
	"github.com/kyma-project/cloud-manager/pkg/testinfra/infraScheme"
	"github.com/kyma-project/cloud-manager/pkg/testinfra/infraTypes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	DefaultSkrNamespace    = testinfra.DefaultSkrNamespace
	DefaultKcpNamespace    = testinfra.DefaultKcpNamespace
	DefaultGardenNamespace = testinfra.DefaultGardenNamespace
)

func SetDefaultNamespace(obj client.Object) {
	switch infraScheme.ObjToClusterType(obj) {
	case infraTypes.ClusterTypeKcp:
		obj.SetNamespace(DefaultKcpNamespace)
	case infraTypes.ClusterTypeSkr:
		obj.SetNamespace(DefaultSkrNamespace)
	case infraTypes.ClusterTypeGarden:
		obj.SetNamespace(DefaultGardenNamespace)
	}
}
