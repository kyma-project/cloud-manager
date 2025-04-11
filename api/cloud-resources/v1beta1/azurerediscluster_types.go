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

// +kubebuilder:validation:Enum=C1;C2;C3;C4;C5
type AzureRedisClusterTier string

const (
	AzureRedisTierC1 AzureRedisClusterTier = "C1"
	AzureRedisTierC2 AzureRedisClusterTier = "C2"
	AzureRedisTierC3 AzureRedisClusterTier = "C3"
	AzureRedisTierC4 AzureRedisClusterTier = "C4"
	AzureRedisTierC5 AzureRedisClusterTier = "C5"
)

type RedisClusterAzureConfigs struct {
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
}

// AzureRedisClusterSpec defines the desired state of AzureRedisCluster
type AzureRedisClusterSpec struct {
	// +kubebuilder:validation:Required
	RedisTier AzureRedisClusterTier `json:"redisTier"`

	// +optional
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="RedisConfiguration is immutable."
	RedisConfiguration RedisClusterAzureConfigs `json:"redisConfiguration"`

	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=10
	ShardCount int32 `json:"shardCount,omitempty"`

	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=10
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="ReplicasPerPrimary is immutable."
	ReplicasPerPrimary int32 `json:"replicasPerPrimary,omitempty"`

	// +optional
	RedisVersion string `json:"redisVersion,omitempty"`

	AuthSecret *RedisAuthSecretSpec `json:"volume,omitempty"`

	// +optional
	IpRange IpRangeRef `json:"ipRange"`
}

// AzureRedisClusterStatus defines the observed state of AzureRedisCluster
type AzureRedisClusterStatus struct {

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

// AzureRedisCluster is the Schema for the AzureredisClusters API
type AzureRedisCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AzureRedisClusterSpec   `json:"spec,omitempty"`
	Status AzureRedisClusterStatus `json:"status,omitempty"`
}

func (in *AzureRedisCluster) GetIpRangeRef() IpRangeRef {
	return in.Spec.IpRange
}

func (in *AzureRedisCluster) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

func (in *AzureRedisCluster) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *AzureRedisCluster) SpecificToFeature() featuretypes.FeatureName {
	return featuretypes.FeatureRedisCluster
}

func (in *AzureRedisCluster) SpecificToProviders() []string {
	return []string{"azure"}
}

func (in *AzureRedisCluster) State() string {
	return in.Status.State
}

func (in *AzureRedisCluster) SetState(v string) {
	in.Status.State = v
}

//+kubebuilder:object:root=true

// AzureRedisClusterList contains a list of AzureRedisCluster
type AzureRedisClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AzureRedisCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AzureRedisCluster{}, &AzureRedisClusterList{})
}
