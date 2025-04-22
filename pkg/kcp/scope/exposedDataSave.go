package scope

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func exposedDataSave(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if !composed.IsObjLoaded(ctx, state) {
		return nil, ctx
	}

	state.ObjAsScope().Status.ExposedData.ReadTime = ptr.To(metav1.Now())

	return composed.PatchStatus(state.ObjAsScope()).
		ErrorLogMessage("Error updating exposed data").
		FailedError(composed.StopWithRequeue).
		SuccessErrorNil().
		SuccessLogMsg("Exposed Data updated").
		Run(ctx, state)
}
