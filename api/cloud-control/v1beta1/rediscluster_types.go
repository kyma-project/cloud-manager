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

type RedisClusterGcp struct {
}

type RedisClusterAzure struct {
}

type RedisClusterAws struct {
}

// +kubebuilder:validation:MinProperties=1
// +kubebuilder:validation:MaxProperties=1
type RedisClusterInfo struct {
	// +optional
	Gcp *RedisClusterGcp `json:"gcp,omitempty"`

	// +optional
	Azure *RedisClusterAzure `json:"azure,omitempty"`

	// +optional
	Aws *RedisClusterAws `json:"aws,omitempty"`
}

// RedisClusterSpec defines the desired state of RedisCluster
type RedisClusterSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="RemoteRef is immutable."
	RemoteRef RemoteRef `json:"remoteRef"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(size(self.name) > 0), message="IpRange name must not be empty."
	IpRange IpRangeRef `json:"ipRange"`

	// +kubebuilder:validation:Required
	Scope ScopeRef `json:"scope"`

	// +kubebuilder:validation:Required
	Instance RedisClusterInfo `json:"instance"`
}

// RedisClusterStatus defines the observed state of RedisCluster
type RedisClusterStatus struct {
	State StatusState `json:"state,omitempty"`

	// List of status conditions to indicate the status of a RedisInstance.
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// RedisCluster is the Schema for the redisclusters API
type RedisCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RedisClusterSpec   `json:"spec,omitempty"`
	Status RedisClusterStatus `json:"status,omitempty"`
}

func (in *RedisCluster) ScopeRef() ScopeRef {
	return in.Spec.Scope
}

func (in *RedisCluster) SetScopeRef(scopeRef ScopeRef) {
	in.Spec.Scope = scopeRef
}

func (in *RedisCluster) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

func (in *RedisCluster) State() string {
	return string(in.Status.State)
}

func (in *RedisCluster) SetState(v string) {
	in.Status.State = StatusState(v)
}

func (in *RedisCluster) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *RedisCluster) CloneForPatchStatus() client.Object {
	result := &RedisCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RedisCluster",
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

	return result
}

func (in *RedisCluster) SetStatusStateToReady() {
	in.Status.State = StateReady
}

func (in *RedisCluster) SetStatusStateToError() {
	in.Status.State = StateError
}

// +kubebuilder:object:root=true

// RedisClusterList contains a list of RedisCluster
type RedisClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RedisCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RedisCluster{}, &RedisClusterList{})
}
