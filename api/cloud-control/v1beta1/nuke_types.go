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
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// AutomaticNuke defines if during the Scope reconciliation a Nuke resource will be created
// if Kyma does not exist. For the normal behavior it should be true, so that Nuke could
// clean up the Scope's orphaned resources. But since the majority of tests creates Scope
// but doesn't create Kyma, this flag in test suite setup is set to `false`, to prevent deletion
// of all the Scope resources and allow tests to function
var AutomaticNuke = true

type NukeResourceStatus string

const (
	NukeResourceStatusDiscovered NukeResourceStatus = "Discovered"
	NukeResourceStatusDeleting   NukeResourceStatus = "Deleting"
	NukeResourceStatusDeleted    NukeResourceStatus = "Deleted"
)

// NukeSpec defines the desired state of Nuke
type NukeSpec struct {
	Scope ScopeRef `json:"scope"`
}

// NukeStatus defines the observed state of Nuke
type NukeStatus struct {
	// +optional
	State string `json:"state,omitempty"`

	// List of status conditions to indicate the status of a Peering.
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions"`

	InitializedOn *metav1.Time `json:"initializedOn,omitempty"`

	// +listType=map
	// +listMapKey=kind
	Resources []*NukeStatusKind `json:"resources"`
}

func (s *NukeStatus) GetKindNoCreate(kind string) *NukeStatusKind {
	for _, sk := range s.Resources {
		if sk.Kind == kind {
			return sk
		}
	}
	return nil
}

func (s *NukeStatus) GetKind(kind string, resourceType ResourceType) (*NukeStatusKind, bool) {
	result := s.GetKindNoCreate(kind)
	if result != nil {
		return result, false
	}
	result = &NukeStatusKind{
		Kind:         kind,
		ResourceType: resourceType,
		Objects:      map[string]NukeResourceStatus{},
	}
	s.Resources = append(s.Resources, result)
	return result, true
}

type ResourceType string

const (
	KcpManagedResource ResourceType = "KCP"
	ProviderResource   ResourceType = "Provider"
)

type NukeStatusKind struct {
	Kind    string                        `json:"kind"`
	Objects map[string]NukeResourceStatus `json:"objects"`
	// +optional
	// +kubebuilder:default=KCP
	ResourceType ResourceType `json:"resourceType"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"

// Nuke is the Schema for the nukes API
type Nuke struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NukeSpec   `json:"spec,omitempty"`
	Status NukeStatus `json:"status,omitempty"`
}

func (in *Nuke) ScopeRef() ScopeRef {
	return in.Spec.Scope
}

func (in *Nuke) SetScopeRef(scopeRef ScopeRef) {
	in.Spec.Scope = scopeRef
}

func (in *Nuke) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

func (in *Nuke) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *Nuke) State() string {
	return in.Status.State
}

func (in *Nuke) SetState(v string) {
	in.Status.State = v
}

func (in *Nuke) CloneForPatchStatus() client.Object {
	result := &Nuke{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Nuke",
			APIVersion: GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: in.Namespace,
			Name:      in.Name,
		},
		Status: in.Status,
	}
	if result.Status.Conditions == nil {
		result.Status.Conditions = []metav1.Condition{}
	}
	if result.Status.Resources == nil {
		result.Status.Resources = []*NukeStatusKind{}
	}
	return result
}

// +kubebuilder:object:root=true

// NukeList contains a list of Nuke
type NukeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Nuke `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Nuke{}, &NukeList{})
}

func (nsk *NukeStatusKind) GetResourceType() ResourceType {

	switch nsk.ResourceType {
	case ProviderResource:
		return ProviderResource
	default:
		return KcpManagedResource
	}
}
