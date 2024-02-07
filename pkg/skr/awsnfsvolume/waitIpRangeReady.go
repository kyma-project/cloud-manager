package awsnfsvolume

import (
	"context"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	"time"
)

func waitIpRangeReady(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	isReady := meta.IsStatusConditionTrue(state.ObjAsAwsNfsVolume().Status.Conditions, cloudresourcesv1beta1.ConditionTypeReady)
	if isReady {
		return nil, nil
	}

	logger := composed.LoggerFromCtx(ctx)
	logger.WithValues("IpRange", fmt.Sprintf("%s/%s", state.SkrIpRange.Namespace, state.SkrIpRange.Name)).
		Info("IpRange is not ready, delaying AwsNfsVolume provisioning")

	return composed.StopWithRequeueDelay(10 * time.Second), nil
}
