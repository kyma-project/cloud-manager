package subnet

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/common/leases"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func releaseLease(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	subnet := state.ObjAsGcpSubnet()

	leaseName := GetLeaseName(state.Scope().Name)
	leaseNamespace := subnet.Namespace
	holderName := fmt.Sprintf("%s/%s", subnet.Namespace, subnet.Name)

	err := leases.Release(
		ctx,
		state.Cluster(),
		leaseName,
		leaseNamespace,
		holderName,
	)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}

		return composed.LogErrorAndReturn(err, "Error releasing lease", composed.StopWithRequeueDelay(util.Timing.T100ms()), ctx)
	}
	return nil, nil
}
