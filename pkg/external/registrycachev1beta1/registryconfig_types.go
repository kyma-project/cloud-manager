package registrycachev1beta1

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RegistryCacheConfigSpec defines the desired state of RegistryCacheConfig.
type RegistryCacheConfigSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Upstream is the remote registry host to cache.
	Upstream string `json:"upstream"`
	// RemoteURL is the remote registry URL. The format must be `<scheme><host>[:<port>]` where
	// `<scheme>` is `https://` or `http://` and `<host>[:<port>]` corresponds to the Upstream
	//
	// If defined, the value is set as `proxy.remoteurl` in the registry [configuration](https://github.com/distribution/distribution/blob/main/docs/content/recipes/mirror.md#configure-the-cache)
	// and in containerd configuration as `server` field in [hosts.toml](https://github.com/containerd/containerd/blob/main/docs/hosts.md#server-field) file.
	// +optional
	RemoteURL *string `json:"remoteURL,omitempty"`
	// Volume contains settings for the registry cache volume.
	// +optional
	Volume *Volume `json:"volume,omitempty"`
	// GarbageCollection contains settings for the garbage collection of content from the cache.
	// Defaults to enabled garbage collection.
	// +optional
	GarbageCollection *GarbageCollection `json:"garbageCollection,omitempty"`
	// SecretReferenceName is the name of the reference for the Secret containing the upstream registry credentials.
	// +optional
	SecretReferenceName *string `json:"secretReferenceName,omitempty"`
	// Proxy contains settings for a proxy used in the registry cache.
	// +optional
	Proxy *Proxy `json:"proxy,omitempty"`

	// HTTP contains settings for the HTTP server that hosts the registry cache.
	HTTP *HTTP `json:"http,omitempty"`
}

// Volume contains settings for the registry cache volume.
type Volume struct {
	// Size is the size of the registry cache volume.
	// Defaults to 10Gi.
	// This field is immutable.
	// +optional
	// +default="10Gi"
	Size *resource.Quantity `json:"size,omitempty"`
	// StorageClassName is the name of the StorageClass used by the registry cache volume.
	// This field is immutable.
	// +optional
	StorageClassName *string `json:"storageClassName,omitempty"`
}

// GarbageCollection contains settings for the garbage collection of content from the cache.
type GarbageCollection struct {
	// TTL is the time to live of a blob in the cache.
	// Set to 0s to disable the garbage collection.
	// Defaults to 168h (7 days).
	// +default="168h"
	TTL metav1.Duration `json:"ttl"`
}

// Proxy contains settings for a proxy used in the registry cache.
type Proxy struct {
	// HTTPProxy field represents the proxy server for HTTP connections which is used by the registry cache.
	// +optional
	HTTPProxy *string `json:"httpProxy,omitempty"`
	// HTTPSProxy field represents the proxy server for HTTPS connections which is used by the registry cache.
	// +optional
	HTTPSProxy *string `json:"httpsProxy,omitempty"`
}

// HTTP contains settings for the HTTP server that hosts the registry cache.
type HTTP struct {
	// TLS indicates whether TLS is enabled for the HTTP server of the registry cache.
	// Defaults to true.
	TLS bool `json:"tls,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// RegistryCacheConfig is the Schema for the registrycacheconfigs API.
type RegistryCacheConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RegistryCacheConfigSpec   `json:"spec,omitempty"`
	Status RegistryCacheConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RegistryCacheConfigList contains a list of CustomConfig.
type RegistryCacheConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RegistryCacheConfig `json:"items"`
}

type State string

const (
	ReadyState   State = "Ready"
	ErrorState   State = "Error"
	PendingState State = "Pending"
)

type ConditionType string

const (
	ConditionTypeRegistryCacheValidated  ConditionType = "RegistryCacheValidated"
	ConditionTypeRegistryCacheConfigured ConditionType = "RegistryCacheConfigured"
)

type ConditionReason string

const (
	ConditionReasonRegistryCacheValidated        ConditionReason = "RegistryCacheValidated"
	ConditionReasonRegistryCacheValidationFailed ConditionReason = "RegistryCacheValidationFailed"

	ConditionReasonRegistryCacheConfigured                       ConditionReason = "RegistryCacheConfigured"
	ConditionReasonRegistryCacheExtensionConfigurationFailed     ConditionReason = "RegistryCacheExtensionConfigurationFailed"
	ConditionReasonRegistryCacheGardenClusterConfigurationFailed ConditionReason = "RegistryCacheGardenClusterConfigurationFailed"
	ConditionReasonRegistryCacheGardenClusterCleanupFailed       ConditionReason = "RegistryCacheGardenClusterCFailedCleanupFailed"
)

type RegistryCacheConfigStatus struct {
	// State signifies current state of Runtime
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=Pending;Ready;Terminating;Failed
	State State `json:"state,omitempty"`

	// List of status conditions to indicate the status of a ServiceInstance.
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

func init() {
	SchemeBuilder.Register(&RegistryCacheConfig{}, &RegistryCacheConfigList{})
}

func (rc *RegistryCacheConfig) RegistryCacheConfiguredUpdateStatusPendingUnknown(reason ConditionReason) {
	rc.updateStatusPending(ConditionTypeRegistryCacheConfigured, reason, metav1.ConditionUnknown)
}

func (rc *RegistryCacheConfig) RegistryCacheConfiguredUpdateStatusFailed(reason ConditionReason, errorMessage string) {
	rc.updateStatusFailed(ConditionTypeRegistryCacheConfigured, reason, metav1.ConditionFalse, errorMessage)
}

func (rc *RegistryCacheConfig) RegistryCacheConfiguredUpdateStatusReady(reason ConditionReason) {
	rc.updateStatusReady(ConditionTypeRegistryCacheConfigured, reason, metav1.ConditionTrue)
}

func (rc *RegistryCacheConfig) updateStatusPending(conditionType ConditionType, reason ConditionReason, status metav1.ConditionStatus) {
	rc.Status.State = PendingState

	condition := metav1.Condition{
		Type:   string(conditionType),
		Reason: string(reason),
		Status: status,
	}

	meta.SetStatusCondition(&rc.Status.Conditions, condition)
}

func (rc *RegistryCacheConfig) updateStatusFailed(conditionType ConditionType, reason ConditionReason, status metav1.ConditionStatus, errorMessage string) {

	rc.Status.State = ErrorState

	condition := metav1.Condition{
		Type:    string(conditionType),
		Reason:  string(reason),
		Status:  status,
		Message: errorMessage,
	}

	meta.SetStatusCondition(&rc.Status.Conditions, condition)
}

func (rc *RegistryCacheConfig) updateStatusReady(conditionType ConditionType, reason ConditionReason, status metav1.ConditionStatus) {
	rc.Status.State = ReadyState

	condition := metav1.Condition{
		Type:   string(conditionType),
		Reason: string(reason),
		Status: status,
	}

	meta.SetStatusCondition(&rc.Status.Conditions, condition)
}
