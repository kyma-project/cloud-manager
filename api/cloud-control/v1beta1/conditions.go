package v1beta1

const (
	ConditionTypeError   = "Error"
	ConditionTypeReady   = "Ready"
	ConditionTypeWarning = "Warning"

	ConditionTypeUpdating = "Updating"

	ReasonScopeNotFound = "ScopeNoFound"

	ReasonUnknown            = "Unknown"
	ReasonNotFound           = "NotFound"
	ReasonReady              = "Ready"
	ReasonGcpError           = "GCPError"
	ReasonValidationFailed   = "ValidationFailed"
	ReasonMissingDependency  = "MissingDependency"
	ReasonWaitingDependency  = "WaitingDependency"
	ReasonDeleteWhileUsed    = "DeleteWhileUsed"
	ReasonCloudProviderError = "CloudProviderError"
)
