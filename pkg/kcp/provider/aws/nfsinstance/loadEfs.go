package nfsinstance

import (
	"context"
	"fmt"
	"time"

	efstypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

// efsMatchesTier reports whether fs matches the reconciled NfsInstance name for the given tier.
// Tiers (first non-empty-match wins): 1=CreationToken, 2=cloud-manager name tag, 3=Name tag.
// Matching is on non-empty values only.
func efsMatchesTier(fs *efstypes.FileSystemDescription, tier int, name string) bool {
	if name == "" {
		return false
	}
	switch tier {
	case 1:
		return ptr.Deref(fs.CreationToken, "") == name
	case 2:
		return awsutil.GetEfsTagValue(fs.Tags, common.TagCloudManagerName) == name
	case 3:
		return awsutil.GetEfsTagValue(fs.Tags, "Name") == name
	}
	return false
}

func loadEfs(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	list, err := state.awsClient.DescribeFileSystems(ctx)
	if err != nil {
		logger.Error(err, "Error listing AWS file systems")
		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
	}

	name := state.Obj().GetName()

	// Three-tier match, first non-empty tier wins. Index into list (not range value) so the
	// collected pointers address distinct elements.
	var matches []*efstypes.FileSystemDescription
	for tier := 1; tier <= 3; tier++ {
		matches = matches[:0]
		for i := range list {
			if efsMatchesTier(&list[i], tier, name) {
				matches = append(matches, &list[i])
			}
		}
		if len(matches) > 0 {
			break
		}
	}

	switch len(matches) {
	case 0:
		// nothing found across all tiers; leave state.efs nil so the create path runs
		return nil, ctx
	case 1:
		state.efs = matches[0]
	default:
		// more than one match within the winning tier: only Status.Id can disambiguate
		statusId := state.ObjAsNfsInstance().Status.Id
		if statusId != "" {
			for _, m := range matches {
				if ptr.Deref(m.FileSystemId, "") == statusId {
					state.efs = m
					break
				}
			}
		}
		if state.efs == nil {
			// genuinely ambiguous: park and alarm, never pick arbitrarily, never create
			ids := make([]string, 0, len(matches))
			for _, m := range matches {
				ids = append(ids, ptr.Deref(m.FileSystemId, ""))
			}
			logger.Error(
				fmt.Errorf("multiple EFS match NfsInstance %q with no matching status.id", name),
				"Multiple AWS EFS matched NfsInstance",
				"efsIds", ids,
			)
			return composed.NewStatusPatcherComposed(state.ObjAsNfsInstance()).
				SetConditions(metav1.Condition{
					Type:    cloudcontrolv1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  cloudcontrolv1beta1.ReasonMultipleFileSystems,
					Message: fmt.Sprintf("Multiple file systems match this NfsInstance: %v", ids),
				}).
				OnSuccess(composed.RequeueAfter(time.Hour)).
				OnFailure(
					composed.Log("Error patching NfsInstance status with multiple file systems error"),
					composed.RequeueAfter(time.Hour),
				).
				Run(ctx, state.Cluster().K8sClient())
		}
	}

	if state.efs.SizeInBytes != nil {
		capacity := resource.NewQuantity(state.efs.SizeInBytes.Value, resource.BinarySI)
		if *capacity != state.ObjAsNfsInstance().Status.Capacity {
			state.ObjAsNfsInstance().Status.Capacity = *capacity
			err = state.PatchObjStatus(ctx)
			if err != nil {
				return composed.LogErrorAndReturn(err, "Error patching NfsInstance status with capacity", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
			}
		}
	}

	logger = logger.WithValues(
		"efsId", ptr.Deref(state.efs.FileSystemId, ""),
		"efsLifeCycleState", state.efs.LifeCycleState,
	)
	ctx = composed.LoggerIntoCtx(ctx, logger)

	if len(state.ObjAsNfsInstance().Status.Id) > 0 {
		return nil, ctx
	}

	state.ObjAsNfsInstance().Status.Id = *state.efs.FileSystemId
	err = state.UpdateObjStatus(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating NfsInstance status with file system id", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}

	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
}
