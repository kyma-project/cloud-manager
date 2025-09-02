package infrastructuremanagerv1

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="STATE",type=string,JSONPath=`.status.state`
//+kubebuilder:printcolumn:name="RUNTIME-ID",type=string,JSONPath=`.metadata.labels.kyma-project\.io/runtime-id`
//+kubebuilder:printcolumn:name="SHOOT-NAME",type=string,JSONPath=`.metadata.labels.kyma-project\.io/shoot-name`
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// GardenerCluster is the Schema for the clusters API
type GardenerCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GardenerClusterSpec   `json:"spec"`
	Status GardenerClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GardenerClusterList contains a list of GardenerCluster
type GardenerClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GardenerCluster `json:"items"`
}

// GardenerClusterSpec defines the desired state of GardenerCluster
type GardenerClusterSpec struct {
	Kubeconfig Kubeconfig `json:"kubeconfig"`
	Shoot      Shoot      `json:"shoot"`
}

// Shoot defines the name of the Shoot resource
type Shoot struct {
	Name string `json:"name"`
}

// Kubeconfig defines the desired kubeconfig location
type Kubeconfig struct {
	Secret Secret `json:"secret"`
}

// SecretKeyRef defines the location, and structure of the secret containing kubeconfig
type Secret struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Key       string `json:"key"`
}

type State string

const (
	ReadyState State = "Ready"
	ErrorState State = "Error"
)

type ConditionReason string

const (
	ConditionReasonKubeconfigSecretCreated ConditionReason = "KubeconfigSecretCreated"
	ConditionReasonKubeconfigSecretRotated ConditionReason = "KubeconfigSecretRotated"
	ConditionReasonFailedToGetSecret       ConditionReason = "FailedToCheckSecret"
	ConditionReasonFailedToCreateSecret    ConditionReason = "ConditionReasonFailedToCreateSecret"
	ConditionReasonFailedToDeleteSecret    ConditionReason = "ConditionReasonFailedToDeleteSecret"
	ConditionReasonFailedToUpdateSecret    ConditionReason = "FailedToUpdateSecret"
	ConditionReasonFailedToGetKubeconfig   ConditionReason = "FailedToGetKubeconfig"
)

type ConditionType string

const (
	ConditionTypeKubeconfigManagement ConditionType = "KubeconfigManagement"
)

// GardenerClusterStatus defines the observed state of GardenerCluster
type GardenerClusterStatus struct {
	// State signifies current state of Gardener Cluster.
	// Value can be one of ("Ready", "Processing", "Error", "Deleting").
	State State `json:"state,omitempty"`

	// List of status conditions to indicate the status of a ServiceInstance.
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

func (cluster *GardenerCluster) UpdateConditionForReadyState(conditionType ConditionType, reason ConditionReason, conditionStatus metav1.ConditionStatus) {
	cluster.Status.State = ReadyState

	condition := metav1.Condition{
		Type:               string(conditionType),
		Status:             conditionStatus,
		LastTransitionTime: metav1.Now(),
		Reason:             string(reason),
		Message:            getMessage(reason),
	}
	meta.RemoveStatusCondition(&cluster.Status.Conditions, condition.Type)
	meta.SetStatusCondition(&cluster.Status.Conditions, condition)
}

func (cluster *GardenerCluster) UpdateConditionForErrorState(conditionType ConditionType, reason ConditionReason, error error) {
	cluster.Status.State = ErrorState

	condition := metav1.Condition{
		Type:               string(conditionType),
		Status:             metav1.ConditionFalse,
		LastTransitionTime: metav1.Now(),
		Reason:             string(reason),
		Message:            fmt.Sprintf("%s Error: %s", getMessage(reason), error.Error()),
	}
	meta.RemoveStatusCondition(&cluster.Status.Conditions, condition.Type)
	meta.SetStatusCondition(&cluster.Status.Conditions, condition)
}

func getMessage(reason ConditionReason) string {
	switch reason {
	case ConditionReasonKubeconfigSecretCreated:
		return "Secret created successfully."
	case ConditionReasonKubeconfigSecretRotated:
		return "Secret rotated successfully."
	case ConditionReasonFailedToCreateSecret:
		return "Failed to create secret."
	case ConditionReasonFailedToUpdateSecret:
		return "Failed to rotate secret."
	case ConditionReasonFailedToGetSecret:
		return "Failed to get secret."
	case ConditionReasonFailedToGetKubeconfig:
		return "Failed to get kubeconfig."

	default:
		return "Unknown condition"
	}
}

func init() {
	SchemeBuilder.Register(&GardenerCluster{}, &GardenerClusterList{})
}
