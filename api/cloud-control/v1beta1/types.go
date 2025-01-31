package v1beta1

type StatusState string

const (
	StateReady      StatusState = "Ready"
	StateError      StatusState = "Error"
	StateProcessing StatusState = "Processing"
	StateWarning    StatusState = "Warning"
	StateDeleting   StatusState = "Deleting"
)

const (
	DO_NOT_USE_FinalizerName = "cloud-control.kyma-project.io/deletion-hook"
)
