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

const (
	ServedTrue  = "True"
	ServedFalse = "False"
)

const (
	ReasonOtherIsServed  = "OtherIsServed"
	ReasonResourcesExist = "ResourcesExist"
)

type ModuleState string

// Valid Module CR States.
const (
	// ModuleStateReady signifies Module CR is Ready and has been installed successfully.
	ModuleStateReady ModuleState = "Ready"

	// ModuleStateDeleting signifies Module CR is being deleted. This is the state that is used
	// when a deletionTimestamp was detected and Finalizers are picked up.
	ModuleStateDeleting ModuleState = "Deleting"

	// ModuleStateWarning signifies specified resource has been deployed, but cannot be used due to misconfiguration,
	// usually it means that user interaction is required.
	ModuleStateWarning ModuleState = "Warning"
)

// CloudResourcesSpec defines the desired state of CloudResources
type CloudResourcesSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of CloudResources. Edit cloudresources_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// CloudResourcesStatus defines the observed state of CloudResources
type CloudResourcesStatus struct {
	State ModuleState `json:"state,omitempty"`

	// +kubebuilder:validation:Enum=True;False
	Served string `json:"served,omitempty"`

	// List of status conditions
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CloudResources is the Schema for the cloudresources API
type CloudResources struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CloudResourcesSpec   `json:"spec,omitempty"`
	Status CloudResourcesStatus `json:"status,omitempty"`
}

func (in *CloudResources) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

func (in *CloudResources) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *CloudResources) SpecificToFeature() featuretypes.FeatureName {
	return ""
}

func (in *CloudResources) SpecificToProviders() []string {
	return nil
}

//+kubebuilder:object:root=true

// CloudResourcesList contains a list of CloudResources
type CloudResourcesList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CloudResources `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CloudResources{}, &CloudResourcesList{})
}
