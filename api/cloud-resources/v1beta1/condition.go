package v1beta1

const (
	ConditionTypeSubmitted = "Submitted"

	ConditionReasonSubmissionSucceeded = "SubmissionSucceeded"
	ConditionReasonSubmissionFailed    = "SubmissionFailed"
)

const (
	ConditionTypeReady = "Ready"

	ConditionReasonReady = "Ready"
)

const (
	ConditionTypeError = "Error"

	// ConditionReasonIpRangeNotFound used with ConditionTypeError in case IpRange specified in object does not exist
	ConditionReasonIpRangeNotFound         = "IpRangeNotFound"
	ConditionReasonMissingScope            = "MissingScope"
	ConditionReasonMissingNfsVolume        = "MissingNfsVolume"
	ConditionReasonNfsVolumeNotReady       = "NfsVolumeNotReady"
	ConditionReasonMissingNfsVolumeBackup  = "MissingNfsVolumeBackup"
	ConditionReasonNfsVolumeBackupNotReady = "NfsVolumeBackupNotReady"
	ConditionReasonNfsRestoreInProgress    = "NfsRestoreInProgress"
	ConditionReasonNfsRestoreFailed        = "NfsRestoreFailed"
	ConditionReasonError                   = "Error"
)

const (
	ConditionTypeUpdating = "Updating"
)

const (
	ConditionTypeDeleting = "Deleting"

	ConditionReasonDeletingPV         = "DeletingPersistentVolume"
	ConditionReasonDeletingPVC        = "DeletingPersistentVolumeClaim"
	ConditionReasonDeletingInstance   = "DeletingInstance"
	ConditionReasonDeletingAuthSecret = "DeletingAuthSecret"
)

const (
	ConditionTypeProcessing = "Processing"
)

const (
	ConditionTypeWaitScopeReady = "WaitScopeReady"
)

const (
	ConditionTypeWarning = "Warning"

	ConditionReasonResourcesExist = "ResourcesExist"
)

const (
	ConditionTypeQuotaExceeded = "QuotaExceeded"
)

const (
	ReasonInvalidCronExpression = "InvalidCronExpression"
	ReasonTimeParseError        = "TimeParseError"
	ReasonScheduleError         = "ScheduleError"
	ReasonNfsVolumeNotFound     = "NfsVolumeNotFound"
	ReasonNfsVolumeNotReady     = "NfsVolumeNotReady"
	ReasonBackupCreateFailed    = "BackupCreateFailed"
	ReasonBackupListFailed      = "BackupListFailed"
	ReasonUnknownSchedule       = "UnknownSchedule"
	ReasonInvalidStartTime      = "InvalidStartTime"
	ReasonInvalidEndTime        = "InvalidEndTime"
	ReasonBackupFailed          = "BackupFailed"
)
