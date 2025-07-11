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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +kubebuilder:validation:Enum=C1;C3;C4;C6;
type GcpRedisClusterTier string

const (
	GcpRedisClusterTierC1 GcpRedisClusterTier = "C1"
	GcpRedisClusterTierC3 GcpRedisClusterTier = "C3"
	GcpRedisClusterTierC4 GcpRedisClusterTier = "C4"
	GcpRedisClusterTierC6 GcpRedisClusterTier = "C6"
)

// GcpRedisClusterSpec defines the desired state of GcpRedisCluster
// +kubebuilder:validation:XValidation:rule=(self.replicasPerShard != 0 || self.shardCount <= 250), message="shardCount must be 250 or less when replicasPerShard is 0"
// +kubebuilder:validation:XValidation:rule=(self.replicasPerShard != 1 || self.shardCount <= 125), message="shardCount must be 125 or less when replicasPerShard is 1"
// +kubebuilder:validation:XValidation:rule=(self.replicasPerShard != 2 || self.shardCount <= 83), message="shardCount must be 83 or less when replicasPerShard is 2"
type GcpRedisClusterSpec struct {
	// +optional
	Subnet GcpSubnetRef `json:"subnet"`

	// +kubebuilder:validation:Required
	RedisTier GcpRedisClusterTier `json:"redisTier"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	ShardCount int32 `json:"shardCount"`

	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=2
	ReplicasPerShard int32 `json:"replicasPerShard"`

	// Redis configuration parameters, according to http://redis.io/topics/config.
	// See docs for the list of the supported parameters
	// +optional
	RedisConfigs map[string]string `json:"redisConfigs"`

	// +optional
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="AuthSecret is immutable."
	AuthSecret *RedisAuthSecretSpec `json:"authSecret,omitempty"`
}

// GcpRedisClusterStatus defines the observed state of GcpRedisCluster
type GcpRedisClusterStatus struct {
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

// GcpRedisCluster is the Schema for the gcpredisclusters API
type GcpRedisCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GcpRedisClusterSpec   `json:"spec,omitempty"`
	Status GcpRedisClusterStatus `json:"status,omitempty"`
}

func (in *GcpRedisCluster) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

func (in *GcpRedisCluster) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *GcpRedisCluster) SpecificToFeature() featuretypes.FeatureName {
	return featuretypes.FeatureRedisCluster
}

func (in *GcpRedisCluster) SpecificToProviders() []string {
	return []string{"gcp"}
}

func (in *GcpRedisCluster) GetGcpSubnetRef() GcpSubnetRef {
	return in.Spec.Subnet
}

func (in *GcpRedisCluster) State() string {
	return in.Status.State
}

func (in *GcpRedisCluster) SetState(v string) {
	in.Status.State = v
}

func (in *GcpRedisCluster) CloneForPatchStatus() client.Object {
	return &GcpRedisCluster{
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
