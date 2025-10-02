package infraScheme

import (
	"github.com/kyma-project/cloud-manager/pkg/common/bootstrap"
	"github.com/kyma-project/cloud-manager/pkg/testinfra/infraTypes"
	"k8s.io/apimachinery/pkg/runtime"
)

var SchemeMap map[infraTypes.ClusterType]*runtime.Scheme

func init() {
	SchemeMap = map[infraTypes.ClusterType]*runtime.Scheme{
		infraTypes.ClusterTypeKcp:    bootstrap.KcpScheme,
		infraTypes.ClusterTypeSkr:    bootstrap.SkrScheme,
		infraTypes.ClusterTypeGarden: bootstrap.GardenScheme,
	}
}
