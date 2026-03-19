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

// AwsWebAclSpec defines the desired state of AwsWebAcl
type AwsWebAclSpec struct {
}

// AwsWebAclStatus defines the observed state of AwsWebAcl.
type AwsWebAclStatus struct {

	// List of status conditions to indicate the status of a AwsWebAcl.
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
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"

// AwsWebAcl is the Schema for the awswebacls API
type AwsWebAcl struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// spec defines the desired state of AwsWebAcl
	// +required
	Spec AwsWebAclSpec `json:"spec,omitempty"`

	// status defines the observed state of AwsWebAcl
	// +optional
	Status AwsWebAclStatus `json:"status,omitempty"`
}

func (in *AwsWebAcl) Conditions() *[]metav1.Condition { return &in.Status.Conditions }

func (in *AwsWebAcl) GetObjectMeta() *metav1.ObjectMeta { return &in.ObjectMeta }

func (in *AwsWebAcl) SpecificToFeature() featuretypes.FeatureName {
	return featuretypes.FeatureWAF
}

func (in *AwsWebAcl) SpecificToProviders() []string { return []string{"aws"} }

func (in *AwsWebAcl) State() string {
	return in.Status.State
}
func (in *AwsWebAcl) SetState(v string) {
	in.Status.State = v
}

// +kubebuilder:object:root=true

// AwsWebAclList contains a list of AwsWebAcl
type AwsWebAclList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AwsWebAcl `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AwsWebAcl{}, &AwsWebAclList{})
}
