package k8sport

import (
	"context"
	"fmt"
	composedv2 "github.com/kyma-project/cloud-manager/pkg/composed/v2"
)

func ToCtx(ctx context.Context, port K8sPort) context.Context {
	key := fmt.Sprintf("k8sport_%s", port.ClusterId())
	return context.WithValue(ctx, key, port)
}

func FromCtx(ctx context.Context, clusterId string) K8sPort {
	key := fmt.Sprintf("k8sport_%s", clusterId)
	x := ctx.Value(key)
	return x.(K8sPort)
}

func FromCtxDefaultCluster(ctx context.Context) K8sPort {
	return FromCtx(ctx, composedv2.DefaultClusterID)
}
