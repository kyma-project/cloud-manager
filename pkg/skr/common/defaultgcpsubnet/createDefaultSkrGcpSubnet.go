package defaultgcpsubnet

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createDefaultSkrGcpSubnet(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(State)
	logger := composed.LoggerFromCtx(ctx)

	if state.GetSkrGcpSubnet() != nil {
		return nil, nil
	}

	skrGcpSubnet := &cloudresourcesv1beta1.GcpSubnet{
		ObjectMeta: metav1.ObjectMeta{
			Name: "default",
			Labels: map[string]string{
				"app.kubernetes.io/name":       "default-gcpsubnet",
				"app.kubernetes.io/instance":   "default",
				"app.kubernetes.io/version":    "1.0.0",
				"app.kubernetes.io/component":  "cloud-manager",
				"app.kubernetes.io/part-of":    "kyma",
				"app.kubernetes.io/managed-by": "cloud-manager",
			},
		},
		Spec: cloudresourcesv1beta1.GcpSubnetSpec{
			Cidr: "10.251.0.0/22",
		},
	}

	err := state.Cluster().K8sClient().Create(ctx, skrGcpSubnet)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating default SKR GcpSubnet", composed.StopWithRequeue, ctx)
	}

	logger.Info("Created default SKR GcpSubnet")
	state.SetSkrGcpSubnet(skrGcpSubnet)

	return nil, nil
}
