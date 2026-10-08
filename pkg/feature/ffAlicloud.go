package feature

import (
	"context"
)

const alicloudFlagName = "alicloud"

// Alicloud gates all Alicloud provider code paths. While Alicloud support is
// being rolled out, every routing decision that dispatches to the Alicloud
// provider (scope creation, subscription handling, network reference,
// vpcnetwork reconciler, iprange reconciler) MUST check this flag and skip
// Alicloud handling when it returns false.
//
// Default is false - Alicloud is opt-in per landscape via ff_ga.yaml / ff_edge.yaml
// targeting rules until the implementation reaches GA.
var Alicloud = &alicloudInfo{}

type alicloudInfo struct{}

func (f *alicloudInfo) Value(ctx context.Context) bool {
	return provider.BoolVariation(ctx, alicloudFlagName, false)
}
