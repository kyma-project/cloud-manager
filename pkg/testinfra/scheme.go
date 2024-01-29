package testinfra

import (
	gardenapi "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

var schemeMap map[ClusterType]*runtime.Scheme

func init() {
	schemeMap = map[ClusterType]*runtime.Scheme{
		ClusterTypeKcp:    runtime.NewScheme(),
		ClusterTypeSkr:    runtime.NewScheme(),
		ClusterTypeGarden: runtime.NewScheme(),
	}
	// KCP
	utilruntime.Must(clientgoscheme.AddToScheme(schemeMap[ClusterTypeKcp]))
	utilruntime.Must(cloudcontrolv1beta1.AddToScheme(schemeMap[ClusterTypeKcp]))
	// SKR
	utilruntime.Must(clientgoscheme.AddToScheme(schemeMap[ClusterTypeSkr]))
	utilruntime.Must(cloudresourcesv1beta1.AddToScheme(schemeMap[ClusterTypeSkr]))
	// Garden
	utilruntime.Must(clientgoscheme.AddToScheme(schemeMap[ClusterTypeGarden]))
	utilruntime.Must(gardenapi.AddToScheme(schemeMap[ClusterTypeGarden]))
}
