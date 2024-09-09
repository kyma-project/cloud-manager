package infraScheme

import (
	gardenapi "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/testinfra/infraTypes"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

var SchemeMap map[infraTypes.ClusterType]*runtime.Scheme

func init() {
	SchemeMap = map[infraTypes.ClusterType]*runtime.Scheme{
		infraTypes.ClusterTypeKcp:    runtime.NewScheme(),
		infraTypes.ClusterTypeSkr:    runtime.NewScheme(),
		infraTypes.ClusterTypeGarden: runtime.NewScheme(),
	}
	// KCP
	utilruntime.Must(clientgoscheme.AddToScheme(SchemeMap[infraTypes.ClusterTypeKcp]))
	utilruntime.Must(cloudcontrolv1beta1.AddToScheme(SchemeMap[infraTypes.ClusterTypeKcp]))
	// SKR
	utilruntime.Must(clientgoscheme.AddToScheme(SchemeMap[infraTypes.ClusterTypeSkr]))
	utilruntime.Must(cloudresourcesv1beta1.AddToScheme(SchemeMap[infraTypes.ClusterTypeSkr]))
	// Garden
	utilruntime.Must(clientgoscheme.AddToScheme(SchemeMap[infraTypes.ClusterTypeGarden]))
	utilruntime.Must(gardenapi.AddToScheme(SchemeMap[infraTypes.ClusterTypeGarden]))
}
