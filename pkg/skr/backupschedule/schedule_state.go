package backupschedule

import (
	"time"

	"github.com/gorhill/cronexpr"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

// ScheduleState is the minimal interface the common scheduling actions need.
// Provider states implement this by embedding their own state and exposing these methods.
type ScheduleState interface {
	composed.State
	ObjAsBackupSchedule() BackupSchedule
	GetScheduleCalculator() *ScheduleCalculator
	GetCronExpression() *cronexpr.Expression
	SetCronExpression(expr *cronexpr.Expression)
	GetNextRunTime() time.Time
	SetNextRunTime(t time.Time)
	IsCreateRunCompleted() bool
	SetCreateRunCompleted(v bool)
	IsDeleteRunCompleted() bool
	SetDeleteRunCompleted(v bool)
}
