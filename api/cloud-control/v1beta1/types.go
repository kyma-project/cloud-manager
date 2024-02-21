package v1beta1

type StatusState string

const (
	ReadyState StatusState = "Ready"
	ErrorState StatusState = "Error"
)

const (
	FinalizerName = "cloud-control.kyma-project.io/deletion-hook"
)
