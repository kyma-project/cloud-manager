package nfsinstance

import (
	"context"
	"fmt"
	"strings"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func shareExportRead(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.share == nil {
		return nil, ctx
	}

	if state.share.ExportLocation == "" {
		logger.Info("SAP share with no export locations")
		state.ObjAsNfsInstance().Status.State = cloudcontrolv1beta1.StateError
		return composed.PatchStatus(state.ObjAsNfsInstance()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ConditionTypeError,
				Message: "No export locations available",
			}).
			ErrorLogMessage("Error patching SAP NfsInstance status after no export locations available").
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T10000ms())).
			Run(ctx, state)
	}

	logger = logger.WithValues("exportPath", state.share.ExportLocation)

	logger.Info("SAP share export path")

	h, p, err := parseExportUrl(state.share.ExportLocation)
	if err != nil {
		logger.Error(err, "error parsing SAP share export path")
		state.ObjAsNfsInstance().Status.State = cloudcontrolv1beta1.StateError
		return composed.PatchStatus(state.ObjAsNfsInstance()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ConditionTypeError,
				Message: "Invalid export path",
			}).
			ErrorLogMessage("Error patching SAP NfsInstance status after invalid export path").
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
		ErrorLogMessage("Error patching SAP NfsInstance status with host and path").
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
