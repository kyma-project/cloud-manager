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

// AwsVpcPeeringSpec defines the desired state of AwsVpcPeering
type AwsVpcPeeringSpec struct {

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="RemoteVpcId is immutable."
	RemoteVpcId string `json:"remoteVpcId"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="RemoteRegion is immutable."
	RemoteRegion string `json:"remoteRegion,omitempty"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="RemoteAccountId is immutable."
	RemoteAccountId string `json:"remoteAccountId"`
}

// AwsVpcPeeringStatus defines the observed state of AwsVpcPeering
type AwsVpcPeeringStatus struct {

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

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// AwsVpcPeering is the Schema for the awsvpcpeerings API
type AwsVpcPeering struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AwsVpcPeeringSpec   `json:"spec,omitempty"`
	Status AwsVpcPeeringStatus `json:"status,omitempty"`
}

func (in *AwsVpcPeering) Conditions() *[]metav1.Condition { return &in.Status.Conditions }

func (in *AwsVpcPeering) GetObjectMeta() *metav1.ObjectMeta { return &in.ObjectMeta }

func (in *AwsVpcPeering) SpecificToFeature() featuretypes.FeatureName {
	return featuretypes.FeaturePeering
}

func (in *AwsVpcPeering) SpecificToProviders() []string { return []string{"aws"} }

//+kubebuilder:object:root=true

// AwsVpcPeeringList contains a list of AwsVpcPeering
type AwsVpcPeeringList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AwsVpcPeering `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AwsVpcPeering{}, &AwsVpcPeeringList{})
}
