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

type AuthSecretSpec struct {
	Name        string            `json:"name,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// GcpRedisInstanceSpec defines the desired state of GcpRedisInstance
type GcpRedisInstanceSpec struct {

	// +optional
	IpRange IpRangeRef `json:"ipRange"`

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

	AuthSecret *AuthSecretSpec `json:"volume,omitempty"`
}

// GcpRedisInstanceStatus defines the observed state of GcpRedisInstance
type GcpRedisInstanceStatus struct {

	// +optional
	Id string `json:"id,omitempty"`

	// List of status conditions
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// +optional
	State string `json:"state,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"
// GcpRedisInstance is the Schema for the gcpredisinstances API
type GcpRedisInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GcpRedisInstanceSpec   `json:"spec,omitempty"`
	Status GcpRedisInstanceStatus `json:"status,omitempty"`
}

func (in *GcpRedisInstance) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

func (in *GcpRedisInstance) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *GcpRedisInstance) SpecificToFeature() featuretypes.FeatureName {
	return featuretypes.FeatureRedis
}

func (in *GcpRedisInstance) SpecificToProviders() []string {
	return []string{"gcp"}
}

func (in *GcpRedisInstance) GetIpRangeRef() IpRangeRef {
	return in.Spec.IpRange
}

func (in *GcpRedisInstance) State() string {
	return in.Status.State
}

func (in *GcpRedisInstance) SetState(v string) {
	in.Status.State = v
}

//+kubebuilder:object:root=true

// GcpRedisInstanceList contains a list of GcpRedisInstance
type GcpRedisInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GcpRedisInstance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GcpRedisInstance{}, &GcpRedisInstanceList{})
}
