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

type RedisClusterAzure struct {
	// +kubebuilder:validation:Required
	SKU AzureRedisClusterSKU `json:"sku"`

	// +optional
	RedisConfiguration RedisInstanceAzureConfigs `json:"redisConfiguration"`

	// +optional
	RedisVersion string `json:"redisVersion,omitempty"`

	// +optional
	ShardCount int `json:"shardCount,omitempty"`

	// +optional
	// +kubebuilder:default=0
	// +kubebuilder:validation:XValidation:rule=(oldSelf != 0 || self == 0), message="replicasPerPrimary cannot be added after the cluster creation."
	ReplicasPerPrimary int `json:"replicasPerPrimary,omitempty"`
}

type AzureRedisClusterSKU struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=1;2;3;4;5
	Capacity int `json:"capacity"`
}

type RedisClusterAws struct {
	// +kubebuilder:validation:Required
	CacheNodeType string `json:"cacheNodeType"`

	// +optional
	// +kubebuilder:default="7.0"
	// +kubebuilder:validation:Enum="7.1";"7.0";"6.x"
	// +kubebuilder:validation:XValidation:rule=(self != "7.0" || oldSelf == "7.0" || oldSelf == "6.x"), message="engineVersion cannot be downgraded."
	// +kubebuilder:validation:XValidation:rule=(self != "7.1" || oldSelf == "7.1" || oldSelf == "7.0" || oldSelf == "6.x"), message="engineVersion cannot be downgraded."
	// +kubebuilder:validation:XValidation:rule=(self != "6.x" || oldSelf == "6.x"), message="engineVersion cannot be downgraded."
	EngineVersion string `json:"engineVersion"`

	// +optional
	// +kubebuilder:default=false
	AutoMinorVersionUpgrade bool `json:"autoMinorVersionUpgrade"`

	// +optional
	// +kubebuilder:default=false
	AuthEnabled bool `json:"authEnabled"`

	// Specifies the weekly time range during which maintenance on the cluster is
	// performed. It is specified as a range in the format ddd:hh24:mi-ddd:hh24:mi (24H
	// Clock UTC). The minimum maintenance window is a 60 minute period.
	//
	// Valid values for ddd are: sun mon tue wed thu fri sat
	//
	// Example: sun:23:00-mon:01:30
	// +optional
	PreferredMaintenanceWindow *string `json:"preferredMaintenanceWindow,omitempty"`

	// +optional
	Parameters map[string]string `json:"parameters,omitempty"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=500
	ShardCount int32 `json:"shardCount"`

	// +kubebuilder:default=0
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=5
	ReplicasPerShard int32 `json:"replicasPerShard"`
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
	// +optional
	Id string `json:"id,omitempty"`

	// +optional
	DiscoveryEndpoint string `json:"discoveryEndpoint,omitempty"`

	// +optional
	AuthString string `json:"authString,omitempty"`

	State StatusState `json:"state,omitempty"`

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
