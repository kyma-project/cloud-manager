package v1beta1

const (
	StateReady             = "Ready"
	StateError             = "Error"
	StateWarning           = "Warning"
	StateProcessing        = "Processing"
	StateWaitingScopeReady = "WaitingScopeReady"
	StateCreating          = "Creating"
	StateDeleting          = "Deleting"
	StateUpdating          = "Updating"
	StateDeleted           = "Deleted"
	StateFailed            = "Failed"
)

const (
	JobStateDone       = "Done"
	JobStateFailed     = "Failed"
	JobStateInProgress = "InProgress"
	JobStateProcessing = "Processing"
	JobStateError      = "Error"
	JobStateActive     = "Active"
	JobStateSuspended  = "Suspended"
)
