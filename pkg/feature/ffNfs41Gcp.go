package feature

import (
	"context"
)

const Nfs41GcpFlagName = "nfs41Gcp"

var Nfs41Gcp = &nfs41GcpInfo{}

type nfs41GcpInfo struct{}

func (k *nfs41GcpInfo) Value(ctx context.Context) bool {
	return provider.BoolVariation(ctx, Nfs41GcpFlagName, false)
}
