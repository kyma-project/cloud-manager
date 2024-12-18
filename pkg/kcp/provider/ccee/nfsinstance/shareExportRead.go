package nfsinstance

import (
	"context"
	"fmt"
	"strings"

	"github.com/gophercloud/gophercloud/v2/openstack/sharedfilesystems/v2/shares"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func shareExportRead(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	list, err := state.cceeClient.ListShareExportLocations(ctx, state.share.ID)
	if err != nil {
		logger.Error(err, "error listing CCEE share export locations")
		state.ObjAsNfsInstance().Status.State = cloudcontrolv1beta1.StateError
		return composed.PatchStatus(state.ObjAsNfsInstance()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ConditionTypeError,
				Message: "Error listing CCEE share export locations",
			}).
			ErrorLogMessage("Error patching CCEE NfsInstance status after list export locations failure").
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
			Run(ctx, state)
	}

	if len(list) == 0 {
		logger.Info("CCEE share with no export locations")
		state.ObjAsNfsInstance().Status.State = cloudcontrolv1beta1.StateError
		return composed.PatchStatus(state.ObjAsNfsInstance()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ConditionTypeError,
				Message: "No export locations available",
			}).
			ErrorLogMessage("Error patching CCEE NfsInstance status after no export locations available").
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T10000ms())).
			Run(ctx, state)
	}

	var el *shares.ExportLocation
	for _, x := range list {
		if x.Preferred {
			el = &x
			break
		}
	}
	if el == nil {
		el = &list[0]
	}

	logger = logger.WithValues("exportPath", el.Path)

	logger.Info("CCEE share export path")

	h, p, err := parseExportUrl(el.Path)
	if err != nil {
		logger.Error(err, "error parsing CCEE share export path")
		state.ObjAsNfsInstance().Status.State = cloudcontrolv1beta1.StateError
		return composed.PatchStatus(state.ObjAsNfsInstance()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ConditionTypeError,
				Message: "Invalid export path",
			}).
			ErrorLogMessage("Error patching CCEE NfsInstance status after invalid export path").
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T10000ms())).
			Run(ctx, state)
	}

	if state.ObjAsNfsInstance().Status.Host == h &&
		state.ObjAsNfsInstance().Status.Path == p {
		return nil, nil
	}

	state.ObjAsNfsInstance().Status.Host = h
	state.ObjAsNfsInstance().Status.Path = p

	return composed.PatchStatus(state.ObjAsNfsInstance()).
		ErrorLogMessage("Error patching CCEE NfsInstance status with host and path").
		SuccessErrorNil().
		Run(ctx, state)
}

func parseExportUrl(url string) (host string, path string, err error) {
	chunks := strings.SplitN(url, "/", 2)
	if len(chunks) != 2 {
		return "", "", fmt.Errorf("invalid export path: %s", url)
	}
	host = chunks[0]
	host = strings.TrimSuffix(host, ":")
	path = chunks[1]
	return
}
