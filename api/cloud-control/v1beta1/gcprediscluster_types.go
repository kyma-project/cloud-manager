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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GcpRedisClusterSpec defines the desired state of GcpRedisCluster
type GcpRedisClusterSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="RemoteRef is immutable."
	RemoteRef RemoteRef `json:"remoteRef"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(size(self.name) > 0), message="Subnet name must not be empty."
	Subnet GcpSubnetRef `json:"subnet"`

	// +kubebuilder:validation:Required
	Scope ScopeRef `json:"scope"`

	// The node type determines the sizing and performance of your node.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=REDIS_SHARED_CORE_NANO;REDIS_STANDARD_SMALL;REDIS_HIGHMEM_MEDIUM;REDIS_HIGHMEM_XLARGE
	NodeType string `json:"nodeType"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=125
	ShardCount int32 `json:"shardCount"`

	// +kubebuilder:default=0
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=2
	ReplicasPerShard int32 `json:"replicasPerShard"`

	// Redis configuration parameters, according to http://redis.io/topics/config.
	// See docs for the list of the supported parameters
	// +optional
	RedisConfigs map[string]string `json:"redisConfigs"`
}

// GcpRedisClusterStatus defines the observed state of GcpRedisCluster
type GcpRedisClusterStatus struct {
	// +optional
	Id string `json:"id,omitempty"`

	// +optional
	DiscoveryEndpoint string `json:"discoveryEndpoint,omitempty"`

	// +optional
	AuthString string `json:"authString,omitempty"`

	State StatusState `json:"state,omitempty"`

	// +optional
	CaCert string `json:"caCert,omitempty"`

	// List of status conditions to indicate the status of a RedisInstance.
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Scope",type="string",JSONPath=".spec.scope.name"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"

// GcpRedisCluster is the Schema for the gcpredisclusters API
type GcpRedisCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GcpRedisClusterSpec   `json:"spec,omitempty"`
	Status GcpRedisClusterStatus `json:"status,omitempty"`
}

func (in *GcpRedisCluster) ScopeRef() ScopeRef {
	return in.Spec.Scope
}

func (in *GcpRedisCluster) SetScopeRef(scopeRef ScopeRef) {
	in.Spec.Scope = scopeRef
}

func (in *GcpRedisCluster) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

func (in *GcpRedisCluster) State() string {
	return string(in.Status.State)
}

func (in *GcpRedisCluster) SetState(v string) {
	in.Status.State = StatusState(v)
}

func (in *GcpRedisCluster) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *GcpRedisCluster) CloneForPatchStatus() client.Object {
	result := &GcpRedisCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "GcpRedisCluster",
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

func (in *GcpRedisCluster) SetStatusStateToReady() {
	in.Status.State = StateReady
}

func (in *GcpRedisCluster) SetStatusStateToError() {
	in.Status.State = StateError
}

// +kubebuilder:object:root=true

// GcpRedisClusterList contains a list of GcpRedisCluster
type GcpRedisClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GcpRedisCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GcpRedisCluster{}, &GcpRedisClusterList{})
}
