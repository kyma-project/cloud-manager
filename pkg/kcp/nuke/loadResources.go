package nuke

import (
	"context"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	nuketypes "github.com/kyma-project/cloud-manager/pkg/kcp/nuke/types"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// loadResources lists all KCP kinds and sets them into the state
func loadResources(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	state.Resources = []*nuketypes.ResourceKindState{
		{
			Kind: "VpcPeering",
			List: &cloudcontrolv1beta1.VpcPeeringList{},
		},
		{
			Kind: "RedisInstance",
			List: &cloudcontrolv1beta1.RedisInstanceList{},
		},
		{
			Kind: "NfsInstance",
			List: &cloudcontrolv1beta1.NfsInstanceList{},
		},
		{
			Kind: "IpRange",
			List: &cloudcontrolv1beta1.IpRangeList{},
		},
		{
			Kind: "Network",
			List: &cloudcontrolv1beta1.NetworkList{},
		},
	}

	for _, rks := range state.Resources {
		err := _listResources(
			ctx,
			state.ObjAsNuke().Spec.Scope.Name,
			state.ObjAsNuke().Namespace,
			rks,
			state.Cluster().K8sClient(),
		)
		if err != nil {
			logger.Error(err, "Error listing resources")

			state.ObjAsNuke().Status.State = string(cloudcontrolv1beta1.ErrorState)

			return composed.PatchStatus(state.ObjAsNuke()).
				SetExclusiveConditions(metav1.Condition{
					Type:    cloudcontrolv1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  "ErrorListingResources",
					Message: err.Error(),
				}).
				ErrorLogMessage("Error patching KCP Nuke status after list resources error").
				SuccessError(composed.StopWithRequeueDelay(util.Timing.T300000ms())).
				Run(ctx, state)
		}

	}

	return nil, ctx
}

func _listResources(ctx context.Context, scopeName, namespace string, rks *nuketypes.ResourceKindState, clnt client.Client) error {
	list := rks.List.DeepCopyObject().(client.ObjectList)
	err := clnt.List(ctx, list, client.InNamespace(namespace))
	if meta.IsNoMatchError(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error listing %s resources: %w", rks.Kind, err)
	}
	arr, err := meta.ExtractList(list)
	if err != nil {
		return fmt.Errorf("error extracting list for %s resources: %w", rks.Kind, err)
	}
	for _, obj := range arr {
		x, ok := obj.(focal.CommonObject)
		if !ok {
			return fmt.Errorf("nuke type %s is not a focal.CommonObject", rks.Kind)
		}
		if x.ScopeRef().Name == scopeName {
			rks.Objects = append(rks.Objects, x)
		}
	}

	return nil
}
