package cloudresources

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func deleteUiCms(ctx context.Context, st composed.State) (error, context.Context) {

	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	uiCmList := util.NewUiCmListUnstructured()
	err := state.Cluster().ApiReader().List(ctx, uiCmList)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error listing UI CMs", composed.StopWithRequeue, ctx)
	}

	logger.Info("Checking UI ConfigMaps to uninstall")

	// Loop through the uiCmList
	uiKey := "cloud-manager"

	for _, uiCm := range uiCmList.Items {

		labels := uiCm.GetLabels()
		value, hasUiKey := labels[uiKey]
		if hasUiKey && value == "ui-cm" {
			u := util.NewUiCmUnstructured()
			u.SetName(uiCm.GetName())
			err = state.Cluster().K8sClient().Delete(ctx, u)
			if err != nil {
				return composed.LogErrorAndReturn(err, fmt.Sprintf("Error deleting ConfigMap %s", uiCm.GetName()), composed.StopWithRequeue, ctx)
			}
		} else {
			continue
		}
	}

	return nil, nil
}
