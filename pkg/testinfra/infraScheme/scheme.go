package infraScheme

import (
	commonscheme "github.com/kyma-project/cloud-manager/pkg/common/scheme"
	"github.com/kyma-project/cloud-manager/pkg/testinfra/infraTypes"
	"k8s.io/apimachinery/pkg/runtime"
)

var SchemeMap map[infraTypes.ClusterType]*runtime.Scheme

func init() {
	SchemeMap = map[infraTypes.ClusterType]*runtime.Scheme{
		infraTypes.ClusterTypeKcp:    commonscheme.KcpScheme,
		infraTypes.ClusterTypeSkr:    commonscheme.SkrScheme,
		infraTypes.ClusterTypeGarden: commonscheme.GardenScheme,
	}
}
