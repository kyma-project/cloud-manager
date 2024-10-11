package awsvpcpeering

import (
	"context"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
)

func waitStatusActive(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if !meta.IsStatusConditionTrue(*state.ObjAsAwsVpcPeering().Conditions(), cloudcontrolv1beta1.ConditionTypeReady) ||
		state.ObjAsAwsVpcPeering().Status.State != string(ec2types.VpcPeeringConnectionStateReasonCodeActive) {
		return composed.StopWithRequeueDelay(util.Timing.T1000ms()), nil
	}

	return nil, nil
}
