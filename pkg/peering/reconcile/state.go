package reconcile

import (
	"github.com/kyma-project/cloud-resources-manager/apis/cloud-resources/v1beta1"
	composed "github.com/kyma-project/cloud-resources-manager/pkg/common/composedAction"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ReconcilingState struct {
	composed.BaseState
	NamespacedName types.NamespacedName
	Obj            client.Object
	CloudResources *v1beta1.CloudResources
}
