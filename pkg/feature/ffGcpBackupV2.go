package feature

import (
	"context"
)

const GcpBackupV2FlagName = "gcpBackupV2"

var GcpBackupV2 = &gcpBackupV2Info{}

type gcpBackupV2Info struct{}

func (k *gcpBackupV2Info) Value(ctx context.Context) bool {
	return provider.BoolVariation(ctx, GcpBackupV2FlagName, false)
}
