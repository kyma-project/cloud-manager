package sim

import (
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func HavingRuntimeProvisioningCompleted(obj client.Object) error {
	x, ok := obj.(*infrastructuremanagerv1.Runtime)
	if !ok {
		return fmt.Errorf("expected Runtime type, got: %T", obj)
	}
	if x.Status.ProvisioningCompleted {
		return nil
	}
	return fmt.Errorf("runtime doesnt have ProvisioningCompleted")
}
