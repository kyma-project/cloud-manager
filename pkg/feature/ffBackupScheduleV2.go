package feature

import (
	"context"
)

const BackupScheduleV2FlagName = "backupScheduleV2"

var BackupScheduleV2 = &backupScheduleV2Info{}

type backupScheduleV2Info struct{}

func (k *backupScheduleV2Info) Value(ctx context.Context) bool {
	return provider.BoolVariation(ctx, BackupScheduleV2FlagName, false)
}
