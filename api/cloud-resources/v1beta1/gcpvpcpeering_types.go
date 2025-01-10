package v1beta1

import (
	featuretypes "github.com/kyma-project/cloud-manager/pkg/feature/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// Important: Run "make" to regenerate code after modifying this file

type GcpVpcPeeringSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=false
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="ImportCustomRoutes is immutable."
	ImportCustomRoutes bool `json:"importCustomRoutes,omitempty"`
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="RemotePeeringName is immutable."
	// +kubebuilder:validation:XValidation:rule=(size(self) <= 63 && size(self) >= 1), message="RemotePeeringName should be at least 1 character, with a maximum of 63 characters."
	// +kubebuilder:validation:XValidation:rule=(self.find('^[a-z]([-a-z0-9]*[a-z0-9])?$') != ''), message="RemotePeeringName must start with a lowercase letter, end with a lowercase letter or number, and only contain lowercase letters, numbers, and hyphens."
	RemotePeeringName string `json:"remotePeeringName,omitempty"`
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="RemoteVpc is immutable."
	// +kubebuilder:validation:XValidation:rule=(size(self) <= 63 && size(self) >= 1), message="RemoteVpc should be at least 1 character, with a maximum of 63 characters."
	// +kubebuilder:validation:XValidation:rule=(self.find('^[a-z]([-a-z0-9]*[a-z0-9])?$') != ''), message="RemoteVpc must start with a lowercase letter, end with a lowercase letter or number, and only contain lowercase letters, numbers, and hyphens."
	RemoteVpc string `json:"remoteVpc,omitempty"`
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(size(self) <= 30 && size(self) >= 6), message="RemoteProject must be 6 to 30 characters in length."
	// +kubebuilder:validation:XValidation:rule=(self.find('^[a-z]([-a-z0-9]*[a-z0-9])?$') != ''), message="RemoteProject must start with a lowercase letter, end with a lowercase letter or number, and only contain lowercase letters, numbers, and hyphens."
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="RemoteProject is immutable."
	RemoteProject string `json:"remoteProject,omitempty"`
	// +kubebuilder:default:=false
	DeleteRemotePeering bool `json:"deleteRemotePeering,omitempty"`
}

type GcpVpcPeeringStatus struct {
	// +optional
	Id string `json:"id,omitempty"`
	// List of status conditions to indicate the Peering status.
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// +optional
	State string `json:"state,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={kyma-cloud-manager}
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
