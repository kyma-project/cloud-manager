package v1beta1

const (
	ConditionTypeError = "Error"
	ConditionTypeReady = "Ready"

	ConditionTypeUpdating = "Updating"

	ReasonScopeNotFound = "ScopeNoFound"

	ReasonUnknown           = "Unknown"
	ReasonReady             = "Ready"
	ReasonGcpError          = "GCPError"
	ReasonValidationFailed  = "ValidationFailed"
	ReasonMissingDependency = "MissingDependency"
	ReasonWaitingDependency = "WaitingDependency"
)
