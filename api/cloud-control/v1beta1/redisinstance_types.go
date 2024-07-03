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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RedisInstanceSpec defines the desired state of RedisInstance
type RedisInstanceSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="RemoteRef is immutable."
	RemoteRef RemoteRef `json:"remoteRef"`

	// +kubebuilder:validation:Required
	IpRange IpRangeRef `json:"ipRange"`

	// +kubebuilder:validation:Required
	Scope ScopeRef `json:"scope"`

	// +kubebuilder:validation:Required
	Instance RedisInstanceInfo `json:"instance"`
}

// +kubebuilder:validation:MinProperties=1
// +kubebuilder:validation:MaxProperties=1
type RedisInstanceInfo struct {
	// +optional
	Gcp *RedisInstanceGcp `json:"gcp,omitempty"`

	// +optional
	Azure *RedisInstanceAzure `json:"azure,omitempty"`

	// +optional
	Aws *RedisInstanceAws `json:"aws,omitempty"`
}

type RedisInstanceGcpConfigs struct {
	// +optional
	MaxmemoryPolicy string `json:"maxmemory-policy,omitempty"`
	// +optional
	NotifyKeyspaceEvents string `json:"notify-keyspace-events,omitempty"`

	// +optional
	Activedefrag string `json:"activedefrag,omitempty"`
	// +optional
	LfuDecayTime string `json:"lfu-decay-time,omitempty"`
	// +optional
	LfuLogFactor string `json:"lfu-log-factor,omitempty"`
	// +optional
	MaxmemoryGb string `json:"maxmemory-gb,omitempty"`

	// +optional
	StreamNodeMaxBytes string `json:"stream-node-max-bytes,omitempty"`
	// +optional
	StreamNodeMaxEntries string `json:"stream-node-max-entries,omitempty"`
}

func (redisConfigs *RedisInstanceGcpConfigs) ToMap() map[string]string {
	result := map[string]string{}

	if redisConfigs.MaxmemoryPolicy != "" {
		result["maxmemory-policy"] = redisConfigs.MaxmemoryPolicy
	}
	if redisConfigs.NotifyKeyspaceEvents != "" {
		result["notify-keyspace-events"] = redisConfigs.NotifyKeyspaceEvents
	}

	if redisConfigs.Activedefrag != "" {
		result["activedefrag"] = redisConfigs.Activedefrag
	}
	if redisConfigs.LfuDecayTime != "" {
		result["lfu-decay-time"] = redisConfigs.LfuDecayTime
	}
	if redisConfigs.LfuLogFactor != "" {
		result["lfu-log-factor"] = redisConfigs.LfuLogFactor
	}
	if redisConfigs.MaxmemoryGb != "" {
		result["maxmemory-gb"] = redisConfigs.MaxmemoryGb
	}

	if redisConfigs.StreamNodeMaxBytes != "" {
		result["stream-node-max-bytes"] = redisConfigs.StreamNodeMaxBytes
	}
	if redisConfigs.StreamNodeMaxEntries != "" {
		result["stream-node-max-entries"] = redisConfigs.StreamNodeMaxEntries
	}

	return result
}

type RedisInstanceGcp struct {
	// +kubebuilder:default=BASIC
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="Tier is immutable."
	Tier string `json:"tier"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="MemorySizeGb is immutable."
	MemorySizeGb int32 `json:"memorySizeGb"`

	// +kubebuilder:default=REDIS_7_0
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="RedisVersion is immutable."
	// +kubebuilder:validation:Enum=REDIS_7_0;REDIS_6_X;REDIS_5_0;REDIS_4_0;REDIS_3_2
	RedisVersion string `json:"redisVersion"`

	// +optional
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="RedisConfigs is immutable."
	RedisConfigs RedisInstanceGcpConfigs `json:"redisConfigs"`
}

type RedisInstanceAzure struct {
}

type RedisInstanceAws struct {
}

// RedisInstanceStatus defines the observed state of RedisInstance
type RedisInstanceStatus struct {
	State StatusState `json:"state,omitempty"`

	// +optional
	Id string `json:"id,omitempty"`

	// +optional
	PrimaryEndpoint string `json:"primaryEndpoint,omitempty"`

	// +optional
	ReadEndpoint string `json:"readEndpoint,omitempty"`

	// +optional
	AuthString string `json:"authString,omitempty"`

	// List of status conditions to indicate the status of a RedisInstance.
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// RedisInstance is the Schema for the redisinstances API
type RedisInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RedisInstanceSpec   `json:"spec,omitempty"`
	Status RedisInstanceStatus `json:"status,omitempty"`
}

func (in *RedisInstance) ScopeRef() ScopeRef {
	return in.Spec.Scope
}

func (in *RedisInstance) SetScopeRef(scopeRef ScopeRef) {
	in.Spec.Scope = scopeRef
}

func (in *RedisInstance) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

func (in *RedisInstance) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *RedisInstance) SetStatusStateToReady() {
	in.Status.State = ReadyState
}

func (in *RedisInstance) SetStatusStateToError() {
	in.Status.State = ErrorState
}

//+kubebuilder:object:root=true

// RedisInstanceList contains a list of RedisInstance
type RedisInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RedisInstance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RedisInstance{}, &RedisInstanceList{})
}
