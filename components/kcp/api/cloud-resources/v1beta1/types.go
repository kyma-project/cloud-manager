package v1beta1

type StatusState string

const (
	UnknownState StatusState = "Unknown"
	ReadyState   StatusState = "Ready"
	ErrorState   StatusState = "Error"
)
