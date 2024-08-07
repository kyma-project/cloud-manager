package v1beta1

import (
	featuretypes "github.com/kyma-project/cloud-manager/pkg/feature/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// Important: Run "make" to regenerate code after modifying this file

type GcpVpcPeeringSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="ImportCustomRoutes is immutable."
	ImportCustomRoutes bool `json:"importCustomRoutes,omitempty"`
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="RemotePeeringName is immutable."
	RemotePeeringName string `json:"remotePeeringName,omitempty"`
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="RemoteVpc is immutable."
	RemoteVpc string `json:"remoteVpc,omitempty"`
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="RemoteNetwork is immutable."
	RemoteProject string `json:"remoteProject,omitempty"`
}

type GcpVpcPeeringStatus struct {
	// List of status conditions to indicate the Peering status.
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:subresource:status

type GcpVpcPeering struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              GcpVpcPeeringSpec   `json:"spec,omitempty"`
	Status            GcpVpcPeeringStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type GcpVpcPeeringList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GcpVpcPeering `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GcpVpcPeering{}, &GcpVpcPeeringList{})
}

func (in *GcpVpcPeering) Conditions() *[]metav1.Condition { return &in.Status.Conditions }

func (in *GcpVpcPeering) GetObjectMeta() *metav1.ObjectMeta { return &in.ObjectMeta }

func (in *GcpVpcPeering) SpecificToFeature() featuretypes.FeatureName {
	return featuretypes.FeaturePeering
}

func (in *GcpVpcPeering) SpecificToProviders() []string { return []string{"gcp"} }
