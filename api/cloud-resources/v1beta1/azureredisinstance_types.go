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

type RedisInstanceAzureConfigs struct {
	// +optional
	MaxClients string `json:"maxclients,omitempty"`
	// +optional
	MaxFragmentationMemoryReserved string `json:"maxfragmentationmemory-reserved,omitempty"`
	// +optional
	MaxMemoryDelta string `json:"maxmemory-delta,omitempty"`
	// +optional
	MaxMemoryPolicy string `json:"maxmemory-policy,omitempty"`
	// +optional
	MaxMemoryReserved string `json:"maxmemory-reserved,omitempty"`
	// +optional
	NotifyKeyspaceEvents string `json:"notify-keyspace-events,omitempty"`
	// +optional
	ZonalConfiguration string `json:"zonal-configuration,omitempty"`
}

type RedisAuthSecretSpec struct {
	Name        string            `json:"name,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}
type AzureRedisSKU struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=1;2;3;4
	Capacity int `json:"capacity"`
}

// AzureRedisInstanceSpec defines the desired state of AzureRedisInstance
type AzureRedisInstanceSpec struct {
	// +kubebuilder:validation:Required
	SKU AzureRedisSKU `json:"sku"`

	// +optional
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="RedisConfiguration is immutable."
	RedisConfiguration RedisInstanceAzureConfigs `json:"redisConfiguration"`

	// +optional
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="RedisVersion is immutable."
	// +kubebuilder:default="6.0"
	RedisVersion string `json:"redisVersion,omitempty"`

	// +optional
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="ShardCount is immutable."
	ShardCount int `json:"shardCount,omitempty"`

	// +optional
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="ReplicasPerPrimary is immutable."
	ReplicasPerPrimary int `json:"replicasPerPrimary,omitempty"`

	AuthSecret *RedisAuthSecretSpec `json:"volume,omitempty"`

	// +optional
	IpRange IpRangeRef `json:"ipRange"`
}

// AzureRedisInstanceStatus defines the observed state of AzureRedisInstance
type AzureRedisInstanceStatus struct {

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
// +kubebuilder:resource:categories={kyma-cloud-manager}
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"

// AzureRedisInstance is the Schema for the Azureredisinstances API
type AzureRedisInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AzureRedisInstanceSpec   `json:"spec,omitempty"`
	Status AzureRedisInstanceStatus `json:"status,omitempty"`
}

func (in *AzureRedisInstance) GetIpRangeRef() IpRangeRef {
	return in.Spec.IpRange
}

func (in *AzureRedisInstance) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

func (in *AzureRedisInstance) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *AzureRedisInstance) SpecificToFeature() featuretypes.FeatureName {
	return featuretypes.FeatureRedis
}

func (in *AzureRedisInstance) SpecificToProviders() []string {
	return []string{"azure"}
}

func (in *AzureRedisInstance) State() string {
	return in.Status.State
}

func (in *AzureRedisInstance) SetState(v string) {
	in.Status.State = v
}

//+kubebuilder:object:root=true

// AzureRedisInstanceList contains a list of AzureRedisInstance
type AzureRedisInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AzureRedisInstance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AzureRedisInstance{}, &AzureRedisInstanceList{})
}
