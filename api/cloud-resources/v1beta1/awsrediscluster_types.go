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

// +kubebuilder:validation:Enum=C1;C2;C3;C4;C5;C6;C7;C8;
type AwsRedisClusterTier string

const (
	AwsRedisTierC1 AwsRedisClusterTier = "C1"
	AwsRedisTierC2 AwsRedisClusterTier = "C2"
	AwsRedisTierC3 AwsRedisClusterTier = "C3"
	AwsRedisTierC4 AwsRedisClusterTier = "C4"
	AwsRedisTierC5 AwsRedisClusterTier = "C5"
	AwsRedisTierC6 AwsRedisClusterTier = "C6"
	AwsRedisTierC7 AwsRedisClusterTier = "C7"
	AwsRedisTierC8 AwsRedisClusterTier = "C8"
)

// AwsRedisClusterSpec defines the desired state of AwsRedisCluster
type AwsRedisClusterSpec struct {
	// +optional
	IpRange IpRangeRef `json:"ipRange"`

	// +optional
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="AuthSecret is immutable."
	AuthSecret *RedisAuthSecretSpec `json:"authSecret,omitempty"`

	// +kubebuilder:validation:Required
	RedisTier AwsRedisClusterTier `json:"redisTier"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=500
	ShardCount int32 `json:"shardCount"`

	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=5
	ReplicasPerShard int32 `json:"replicasPerShard"`

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
}

// AwsRedisClusterStatus defines the observed state of AwsRedisCluster
type AwsRedisClusterStatus struct {
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

// AwsRedisCluster is the Schema for the awsredisclusters API
type AwsRedisCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AwsRedisClusterSpec   `json:"spec,omitempty"`
	Status AwsRedisClusterStatus `json:"status,omitempty"`
}

func (in *AwsRedisCluster) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

func (in *AwsRedisCluster) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *AwsRedisCluster) SpecificToFeature() featuretypes.FeatureName {
	return featuretypes.FeatureRedisCluster
}

func (in *AwsRedisCluster) SpecificToProviders() []string {
	return []string{"aws"}
}

func (in *AwsRedisCluster) GetIpRangeRef() IpRangeRef {
	return in.Spec.IpRange
}

func (in *AwsRedisCluster) State() string {
	return in.Status.State
}

func (in *AwsRedisCluster) SetState(v string) {
	in.Status.State = v
}

func (in *AwsRedisCluster) CloneForPatchStatus() client.Object {
	return &AwsRedisCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AwsRedisCluster",
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

// AwsRedisClusterList contains a list of AwsRedisCluster
type AwsRedisClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AwsRedisCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AwsRedisCluster{}, &AwsRedisClusterList{})
}
