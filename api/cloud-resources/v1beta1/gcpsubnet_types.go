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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GcpSubnetSpec defines the desired state of GcpSubnet
type GcpSubnetSpec struct {
	// +kubebuilder:validation:Required
	Cidr string `json:"cidr"`
}

// GcpSubnetStatus defines the observed state of GcpSubnet
type GcpSubnetStatus struct {
	State string `json:"state,omitempty"`

	// +optional
	Cidr string `json:"cidr,omitempty"`

	// +optional
	Id string `json:"id,omitempty"`

	// List of status conditions
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,categories={kyma-cloud-manager}
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="CIDR",type="string",JSONPath=".spec.cidr"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"

// GcpSubnet is the Schema for the gcpsubnets API
type GcpSubnet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GcpSubnetSpec   `json:"spec,omitempty"`
	Status GcpSubnetStatus `json:"status,omitempty"`
}

func (in *GcpSubnet) State() string {
	return in.Status.State
}

func (in *GcpSubnet) SetState(v string) {
	in.Status.State = v
}

func (in *GcpSubnet) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

func (in *GcpSubnet) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *GcpSubnet) SpecificToFeature() featuretypes.FeatureName {
	return featuretypes.FeatureRedisCluster // will be moved to undefined feature after GcpRedisCluster GA
}

func (in *GcpSubnet) SpecificToProviders() []string {
	return []string{"gcp"}
}

// +kubebuilder:object:root=true

// GcpSubnetList contains a list of GcpSubnet
type GcpSubnetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GcpSubnet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GcpSubnet{}, &GcpSubnetList{})
}
