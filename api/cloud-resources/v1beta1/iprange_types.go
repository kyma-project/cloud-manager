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
	"github.com/elliotchance/pie/v2"
	featuretypes "github.com/kyma-project/cloud-manager/pkg/feature/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ConditionReasonInvalidCidr = "InvalidCidr"
	ConditionReasonCidrOverlap = "CidrOverlap"

	ConditionReasonCidrCanNotBeChanged = "CidrCanNotBeChanged"

	ConditionTypeDeleteWhileUsed = "DeleteWhileUsed"
)

// IpRangeSpec defines the desired state of IpRange
type IpRangeSpec struct {
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="CIDR is immutable"
	// +optional
	Cidr string `json:"cidr"`
}

// IpRangeStatus defines the observed state of IpRange
type IpRangeStatus struct {
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

// IpRange is the Schema for the ipranges API
type IpRange struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IpRangeSpec   `json:"spec,omitempty"`
	Status IpRangeStatus `json:"status,omitempty"`
}

func (in *IpRange) State() string {
	return in.Status.State
}

func (in *IpRange) SetState(v string) {
	in.Status.State = v
}

func (in *IpRange) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

func (in *IpRange) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *IpRange) SpecificToFeature() featuretypes.FeatureName {
	return ""
}

func (in *IpRange) SpecificToProviders() []string {
	return nil
}

//+kubebuilder:object:root=true

// IpRangeList contains a list of IpRange
type IpRangeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IpRange `json:"items"`
}

func (l *IpRangeList) GetItemCount() int {
	return len(l.Items)
}

func (l *IpRangeList) GetItems() []client.Object {
	return pie.Map(l.Items, func(item IpRange) client.Object {
		return &item
	})
}

func init() {
	SchemeBuilder.Register(&IpRange{}, &IpRangeList{})
}
