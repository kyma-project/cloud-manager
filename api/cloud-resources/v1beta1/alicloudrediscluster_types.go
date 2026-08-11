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
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// AlicloudRedisClusterSpec defines the desired state of AlicloudRedisCluster.
//
// An AlicloudRedisCluster provisions a sharded AliCloud r-kvstore cloud-native
// cluster instance. See AlicloudRedisInstance for standard (non-sharded) HA.
type AlicloudRedisClusterSpec struct {
	// IpRange references the IpRange resource used for private connectivity.
	// If omitted, the SKR controller selects the default IpRange.
	// +optional
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="IpRange is immutable."
	IpRange IpRangeRef `json:"ipRange,omitempty"`
	// Note: omitempty above is required. Without it an unset IpRange serialises as {"name":""},
	// which combined with the self==oldSelf XValidation permanently locks IpRange to empty-string.

	// RedisTier defines per-shard capacity.
	// Total capacity = tier_memory × shardCount.
	// +kubebuilder:validation:Required
	RedisTier AlicloudRedisClusterTier `json:"redisTier"`

	// ShardCount is the number of data shards. Mutable via ModifyShardingNum.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=32
	ShardCount int32 `json:"shardCount"`

	// ReplicasPerShard is the number of read-only replicas per shard.
	// All tiers currently exposed via redisTier map to proxy-based classes
	// (redis.logic.sharding.*), which always have ReadOnlyCount=0. This field
	// is reserved for future non-proxy tiers; it must be 0 for all current tiers.
	// +optional
	// +kubebuilder:default=0
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=1
	// +kubebuilder:validation:XValidation:rule=(self == 0), message="replicasPerShard must be 0; all current redisTier values use proxy-based classes that do not support read replicas."
	ReplicasPerShard int32 `json:"replicasPerShard,omitempty"`

	// EngineVersion is the Redis engine version. Immutable after creation.
	// +kubebuilder:default="5.0"
	// +kubebuilder:validation:Enum="5.0";"6.0";"7.0"
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="engineVersion is immutable."
	EngineVersion string `json:"engineVersion,omitempty"`

	// +optional
	AuthSecret *RedisAuthSecretSpec `json:"authSecret,omitempty"`
}

// AlicloudRedisClusterStatus defines the observed state of AlicloudRedisCluster.
type AlicloudRedisClusterStatus struct {
	// +optional
	Id string `json:"id,omitempty"`

	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

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

// AlicloudRedisCluster is the Schema for the alicloudredisclusters API
type AlicloudRedisCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AlicloudRedisClusterSpec   `json:"spec,omitempty"`
	Status AlicloudRedisClusterStatus `json:"status,omitempty"`
}

func (in *AlicloudRedisCluster) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

func (in *AlicloudRedisCluster) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *AlicloudRedisCluster) SpecificToFeature() featuretypes.FeatureName {
	return featuretypes.FeatureRedisCluster
}

func (in *AlicloudRedisCluster) SpecificToProviders() []string {
	return []string{"alicloud"}
}

func (in *AlicloudRedisCluster) GetIpRangeRef() IpRangeRef {
	return in.Spec.IpRange
}

func (in *AlicloudRedisCluster) State() string {
	return in.Status.State
}

func (in *AlicloudRedisCluster) SetState(v string) {
	in.Status.State = v
}

func (in *AlicloudRedisCluster) CloneForPatchStatus() client.Object {
	return &AlicloudRedisCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AlicloudRedisCluster",
			APIVersion: GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: in.Namespace,
			Name:      in.Name,
		},
		Status: in.Status,
	}
}

//+kubebuilder:object:root=true

// AlicloudRedisClusterList contains a list of AlicloudRedisCluster
type AlicloudRedisClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AlicloudRedisCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AlicloudRedisCluster{}, &AlicloudRedisClusterList{})
}
