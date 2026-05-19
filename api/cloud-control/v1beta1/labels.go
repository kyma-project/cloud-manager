package v1beta1

// Used for K8s labels
const (
	LabelKymaName        = "cloud-manager.kyma-project.io/kymaName"
	LabelRemoteName      = "cloud-manager.kyma-project.io/remoteName"
	LabelRemoteNamespace = "cloud-manager.kyma-project.io/remoteNamespace"
)

const (
	LabelIgnore = "cloud-manager.kyma-project.io/ignore"
)

const (
	RuntimeSecurityStatusAnnotation             = "cloud-manager.kyma-project.io/security-status"
	RuntimeSecurityMessageAnnotation            = "cloud-manager.kyma-project.io/security-message"
	RuntimeSecurityObservedGenerationAnnotation = "cloud-manager.kyma-project.io/security-observed-generation"
	RuntimeSecurityLastReconcileTime            = "cloud-manager.kyma-project.io/security-last-reconcile-time"
)
