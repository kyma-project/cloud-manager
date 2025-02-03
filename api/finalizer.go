package api

const (
	CommonFinalizerDeletionHook = "cloud-manager.kyma-project.io/deletion-hook"
)

// TODO: Remove in next release - after 1.2.5 is released, aka in the 1.2.6
// Finalizer migration
const (
	DO_NOT_USE_OLD_KcpFinalizer = "cloud-control.kyma-project.io/deletion-hook"
	DO_NOT_USE_OLD_SkrFinalizer = "cloud-resources.kyma-project.io/deletion-hook"
)
