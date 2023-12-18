package v1beta1

type StatusState string

const (
	UnknownState StatusState = "Unknown"
	ReadyState   StatusState = "Ready"
	ErrorState   StatusState = "Error"
)

type ConditionType string

const (
	ConditionTypeReady      ConditionType = "Ready"
	ConditionTypeProcessing ConditionType = "Processing"
	ConditionTypeDeleted    ConditionType = "Deleted"
	ConditionTypeError      ConditionType = "Error"
)

type ConditionReason string

const (
	ConditionReasonReady      ConditionReason = "Ready"
	ConditionReasonProcessing ConditionReason = "Processing"
	ConditionReasonError      ConditionReason = "Error"
	ConditionReasonDeleted    ConditionReason = "Deleted"
)
