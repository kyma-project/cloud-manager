/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	featuretypes "github.com/kyma-project/cloud-manager/pkg/feature/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AzureVpcPeeringSpec defines the desired state of AzureVpcPeering
type AzureVpcPeeringSpec struct {

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="RemotePeeringName is immutable."
	// +kubebuilder:validation:XValidation:rule=(size(self) <= 80), message="RemotePeeringName can be up to 80 characters long."
	// +kubebuilder:validation:XValidation:rule=(self.find('^[a-z0-9][a-z0-9-]*[a-z0-9]$') != ''), message="RemotePeeringName must begin with a word character, and it must end with a word character. RemotePeeringName may contain word characters or '-'."
	RemotePeeringName string `json:"remotePeeringName,omitempty"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="RemoteVnet is immutable."
	RemoteVnet string `json:"remoteVnet,omitempty"`

	DeleteRemotePeering bool `json:"deleteRemotePeering,omitempty"`

	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="RemoteTenant is immutable."
	RemoteTenant string `json:"remoteTenant,omitempty"`
}

// AzureVpcPeeringStatus defines the observed state of AzureVpcPeering
type AzureVpcPeeringStatus struct {

	// +optional
	Id string `json:"id,omitempty"`

	// List of status conditions to indicate the status of a Peering.
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// +optional
	State string `json:"state,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,categories={kyma-cloud-manager}
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"

// AzureVpcPeering is the Schema for the azurevpcpeerings API
type AzureVpcPeering struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AzureVpcPeeringSpec   `json:"spec,omitempty"`
	Status AzureVpcPeeringStatus `json:"status,omitempty"`
}

func (in *AzureVpcPeering) Conditions() *[]metav1.Condition { return &in.Status.Conditions }

func (in *AzureVpcPeering) GetObjectMeta() *metav1.ObjectMeta { return &in.ObjectMeta }

func (in *AzureVpcPeering) SpecificToFeature() featuretypes.FeatureName {
	return featuretypes.FeaturePeering
}

func (in *AzureVpcPeering) SpecificToProviders() []string { return []string{"azure"} }

func (in *AzureVpcPeering) State() string {
	return in.Status.State
}
func (in *AzureVpcPeering) SetState(v string) {
	in.Status.State = v
}

//+kubebuilder:object:root=true

// AzureVpcPeeringList contains a list of AzureVpcPeering
type AzureVpcPeeringList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AzureVpcPeering `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AzureVpcPeering{}, &AzureVpcPeeringList{})
}
