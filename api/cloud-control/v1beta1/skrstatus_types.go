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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SkrStatusCondition struct {
	Title           string   `json:"title"`
	ObjKindGroup    string   `json:"objKindGroup"`
	CrdKindGroup    string   `json:"crdKindGroup"`
	BusolaKindGroup string   `json:"busolaKindGroup"`
	Feature         string   `json:"feature"`
	ObjName         string   `json:"objName"`
	ObjNamespace    string   `json:"objNamespace"`
	Filename        string   `json:"filename"`
	Ok              bool     `json:"ok"`
	Outcomes        []string `json:"outcomes"`
}

// SkrStatusSpec defines the desired state of SkrStatus.
type SkrStatusSpec struct {
	KymaName      string `json:"kymaName"`
	Provider      string `json:"provider"`
	BrokerPlan    string `json:"brokerPlan"`
	GlobalAccount string `json:"globalAccount"`
	SubAccount    string `json:"subAccount"`
	Region        string `json:"region"`
	ShootName     string `json:"shootName"`

	PastConnections        []metav1.Time `json:"pastConnections,omitempty"`
	AverageIntervalSeconds int           `json:"averageIntervalSeconds,omitempty"`

	// +optional
	Conditions []SkrStatusCondition `json:"conditions"`
}

// SkrStatusStatus defines the observed state of SkrStatus.
type SkrStatusStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// SkrStatus is the Schema for the skrstatuses API.
type SkrStatus struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SkrStatusSpec   `json:"spec,omitempty"`
	Status SkrStatusStatus `json:"status,omitempty"`
}

func (in *SkrStatus) CloneForPatch() *SkrStatus {
	result := &SkrStatus{
		TypeMeta: in.TypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Namespace: in.Namespace,
			Name:      in.Name,
		},
		Spec: in.Spec,
	}
	if len(in.Labels) > 0 {
		result.Labels = in.Labels
	}
	if len(in.Annotations) > 0 {
		result.Annotations = in.Annotations
	}
	return result
}

// +kubebuilder:object:root=true

// SkrStatusList contains a list of SkrStatus.
type SkrStatusList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SkrStatus `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SkrStatus{}, &SkrStatusList{})
}
