package cloudresources

import (
	"context"
	"fmt"
	"strings"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func deleteCrds(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.ObjAsCloudResources().Status.Served != cloudresourcesv1beta1.ServedTrue {
		return nil, nil
	}

	crdList := util.NewCrdListUnstructured()
	err := state.Cluster().ApiReader().List(ctx, crdList)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error listing CRDs", composed.StopWithRequeue, ctx)
	}

	state.IgnoreWatchErrors(true)

	logger.Info("Checking CRDs to uninstall")

	suffix := ".cloud-resources.kyma-project.io"
	skip := "cloudresources.cloud-resources.kyma-project.io"
	for _, crd := range crdList.Items {
		if !strings.HasSuffix(crd.GetName(), suffix) {
			continue
		}
		if crd.GetName() == skip {
			continue
		}

		logger.
			WithValues("crd", crd.GetName()).
			Info("Deleting CRD")

		u := util.NewCrdUnstructured()
		u.SetName(crd.GetName())
		err = state.Cluster().K8sClient().Delete(ctx, u)
		if err != nil {
			return composed.LogErrorAndReturn(err, fmt.Sprintf("Error deleting CRD %s", crd.GetName()), composed.StopWithRequeue, ctx)
		}
	}

	return nil, ctx
}
