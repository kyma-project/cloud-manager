package bootstrap

import (
	gardenapi "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/external/keb"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

var (
	KcpScheme    = runtime.NewScheme()
	SkrScheme    = runtime.NewScheme()
	GardenScheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(GardenScheme))
	utilruntime.Must(gardenapi.AddToScheme(GardenScheme))

	utilruntime.Must(clientgoscheme.AddToScheme(KcpScheme))
	utilruntime.Must(cloudcontrolv1beta1.AddToScheme(KcpScheme))
	utilruntime.Must(apiextensions.AddToScheme(KcpScheme))
	utilruntime.Must(keb.AddToScheme(KcpScheme))

	utilruntime.Must(clientgoscheme.AddToScheme(SkrScheme))
	utilruntime.Must(cloudresourcesv1beta1.AddToScheme(SkrScheme))
	utilruntime.Must(apiextensions.AddToScheme(SkrScheme))
}
