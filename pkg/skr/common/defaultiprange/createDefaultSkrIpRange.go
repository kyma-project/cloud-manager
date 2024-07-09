package defaultiprange

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createDefaultSkrIpRange(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(State)
	logger := composed.LoggerFromCtx(ctx)

	if composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, nil
	}

	if state.GetSkrIpRange() != nil {
		return nil, nil
	}

	skrIpRange := &cloudresourcesv1beta1.IpRange{
		ObjectMeta: metav1.ObjectMeta{
			Name: "default",
			Labels: map[string]string{
				"app.kubernetes.io/name":       "default-iprange",
				"app.kubernetes.io/instance":   "default",
				"app.kubernetes.io/version":    "1.0.0",
				"app.kubernetes.io/component":  "cloud-manager",
				"app.kubernetes.io/part-of":    "kyma",
				"app.kubernetes.io/managed-by": "cloud-manager",
			},
		},
	}

	err := state.Cluster().K8sClient().Create(ctx, skrIpRange)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating default SKR IpRange", composed.StopWithRequeue, ctx)
	}

	logger.Info("Created default SKR IpRange")
	state.SetSkrIpRange(skrIpRange)

	return nil, nil
}
