package cloudresources

import (
	"context"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func handleServed(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	logger.Info("Handling served")

	served, err := func() (*cloudresourcesv1beta1.CloudResources, error) {
		list := &cloudresourcesv1beta1.CloudResourcesList{}
		err := state.Cluster().K8sClient().List(ctx, list)
		if err != nil {
			return nil, err
		}

		var servedCloudResources *cloudresourcesv1beta1.CloudResources
		for _, item := range list.Items {
			if item.Status.Served == cloudresourcesv1beta1.ServedTrue {
				servedCloudResources = &item
				break
			}
		}

		return servedCloudResources, nil
	}()
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error listing CloudResources", composed.StopWithRequeue, ctx)
	}

	if served != nil && state.Obj().GetName() == served.Name && state.Obj().GetNamespace() == served.Namespace {
		// nothing to do, the obj is the served one
		logger.Info("Found served CloudResources")
		return nil, nil
	}

	if state.ObjAsCloudResources().Status.Served == cloudresourcesv1beta1.ServedFalse {
		logger.Info("Ignoring not served CloudResources")
		// we're not handling not served objects
		return composed.StopAndForget, nil
	}

	if served == nil {
		// none is served so far, this obj will be the one
		logger.Info("Setting CloudResources Served to True")
		state.ObjAsCloudResources().Status.Served = cloudresourcesv1beta1.ServedTrue
		return composed.UpdateStatus(state.ObjAsCloudResources()).
			ErrorLogMessage("Error updating CloudResources served to true").
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	logger.Info("Setting CloudResources Served to False")
	state.ObjAsCloudResources().Status.Served = cloudresourcesv1beta1.ServedFalse
	state.ObjAsCloudResources().Status.State = cloudresourcesv1beta1.ModuleStateWarning
	return composed.UpdateStatus(state.ObjAsCloudResources()).
		SetExclusiveConditions(metav1.Condition{
			Type:   cloudresourcesv1beta1.ConditionTypeError,
			Status: metav1.ConditionTrue,
			Reason: cloudresourcesv1beta1.ReasonOtherIsServed,
			Message: fmt.Sprintf("only one instance of CloudResources is allowed (current served instance: %s",
				served.Name),
		}).
		ErrorLogMessage("Error updating CloudResources served to false").
		SuccessError(composed.StopAndForget).
		Run(ctx, state)
}
