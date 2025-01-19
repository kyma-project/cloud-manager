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

type SkrResourceUpgrade struct {
	Date        metav1.Time `json:"date"`
	Name        string      `json:"name"`
	FromVersion string      `json:"fromVersion"`
	ToVersion   string      `json:"toVersion"`
}

type SkrStatusToggles struct {
	ResourceUpgrades    []SkrResourceUpgrade `json:"resourceUpgrades"`
	EnabledControllers  []string             `json:"enabledControllers,omitempty"`
	DisabledControllers []string             `json:"disabledControllers,omitempty"`
	EnabledIndexers     []string             `json:"enabledIndexers,omitempty"`
	DisabledIndexers    []string             `json:"disabledIndexers,omitempty"`
}

type SkrConnectInfo struct {
	Date   metav1.Time `json:"date"`
	Status string      `json:"status"`
}

type SkrConnections struct {
	History                []SkrConnectInfo `json:"history,omitempty"`
	AverageIntervalSeconds int              `json:"averageIntervalSeconds,omitempty"`
}

// SkrStatusSpec defines the desired state of SkrStatus.
type SkrStatusSpec struct {
	Toggles     SkrStatusToggles `json:"toggles,omitempty"`
	Connections SkrConnections   `json:"connections,omitempty"`
	Provider    string           `json:"provider,omitempty"`
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

func (in *SkrStatus) NotReady() {

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
