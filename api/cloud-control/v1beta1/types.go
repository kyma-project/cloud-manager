package v1beta1

type StatusState string

const (
	ReadyState      StatusState = "Ready"
	ErrorState      StatusState = "Error"
	ProcessingState StatusState = "Processing"
	WarningState    StatusState = "Warning"
	DeletingState   StatusState = "Deleting"
)

const (
	FinalizerName = "cloud-control.kyma-project.io/deletion-hook"
)
