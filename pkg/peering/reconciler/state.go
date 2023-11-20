package reconciler

import (
	"github.com/kyma-project/cloud-resources-manager/apis/cloud-resources/v1beta1"
	composed "github.com/kyma-project/cloud-resources-manager/pkg/common/composedAction"
	"k8s.io/apimachinery/pkg/runtime"
)

type ReconcilingState struct {
	composed.BaseState
	Obj            runtime.Object
	CloudResources *v1beta1.CloudResources
}
