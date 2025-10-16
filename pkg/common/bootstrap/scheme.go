package bootstrap

import (
	gardenerapicore "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	gardenerapisecurity "github.com/gardener/gardener/pkg/apis/security/v1alpha1"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"
	"github.com/kyma-project/cloud-manager/pkg/external/operatorv1beta2"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
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
	utilruntime.Must(gardenerapicore.AddToScheme(GardenScheme))
	utilruntime.Must(gardenerapisecurity.AddToScheme(GardenScheme))
	utilruntime.Must(clientgoscheme.AddToScheme(GardenScheme))

	utilruntime.Must(cloudcontrolv1beta1.AddToScheme(KcpScheme))
	utilruntime.Must(apiextensions.AddToScheme(KcpScheme))
	utilruntime.Must(apiextensionsv1.AddToScheme(KcpScheme))
	utilruntime.Must(infrastructuremanagerv1.AddToScheme(KcpScheme))
	utilruntime.Must(operatorv1beta2.AddToScheme(KcpScheme))
	utilruntime.Must(clientgoscheme.AddToScheme(KcpScheme))

	utilruntime.Must(cloudresourcesv1beta1.AddToScheme(SkrScheme))
	utilruntime.Must(apiextensions.AddToScheme(SkrScheme))
	utilruntime.Must(apiextensionsv1.AddToScheme(SkrScheme))
	utilruntime.Must(clientgoscheme.AddToScheme(SkrScheme))
	utilruntime.Must(operatorv1beta2.AddToScheme(SkrScheme))
}
