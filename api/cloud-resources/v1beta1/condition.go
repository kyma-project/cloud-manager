package v1beta1

const (
	ConditionTypeSubmitted = "Submitted"

	ConditionReasonSubmissionSucceeded = "SubmissionSucceeded"
	ConditionReasonSubmissionFailed    = "SubmissionFailed"
)

const (
	ConditionTypeReady = "Ready"

	ConditionReasonError = "Error"
)

const (
	ConditionTypeError = "Error"

	// ConditionReasonIpRangeNotFound used with ConditionTypeError in case IpRange specified in object does not exist
	ConditionReasonIpRangeNotFound   = "IpRangeNotFound"
	ConditionReasonMissingScope      = "MissingScope"
	ConditionReasonMissingNfsVolume  = "MissingNfsVolume"
	ConditionReasonNfsVolumeNotReady = "NfsVolumeNotReady"
)

const (
	ConditionTypeDeleting = "Deleting"

	ConditionReasonDeletingPV       = "DeletingPersistentVolume"
	ConditionReasonDeletingInstance = "DeletingInstance"
)

const (
	ConditionTypeProcessing = "Processing"
)

const (
	ConditionTypeWaitScopeReady = "WaitScopeReady"
)
