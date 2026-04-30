package azurerwxvolumerestore

import (
	"context"

	scopeprovider "github.com/kyma-project/cloud-manager/pkg/skr/common/scope/provider"
)

var scopeProvider = scopeprovider.Always("test-ns", "test-scope")

// Fake client doesn't support type "apply" for patching so falling back on update for unit tests.
func (s *State) PatchObjStatus(ctx context.Context) error {
	return s.Cluster().K8sClient().Status().Update(ctx, s.Obj())
}
