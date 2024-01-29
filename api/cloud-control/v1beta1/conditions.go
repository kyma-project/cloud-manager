package v1beta1

const (
	ConditionTypeError = "Error"
	ConditionTypeReady = "Ready"

	ReasonScopeNotFound = "ScopeNoFound"

	ReasonUnknown          = "Unknown"
	ReasonReady            = "Ready"
	ReasonGcpError         = "GCPError"
	ReasonNotSupported     = "NotSupported"
	ReasonValidationFailed = "ValidationFailed"
)
