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

// AlicloudRedisInstanceSpec defines the desired state of AlicloudRedisInstance.
//
// An AlicloudRedisInstance provisions a standard (non-sharded) HA AliCloud
// r-kvstore instance. See AlicloudRedisCluster for sharded clusters.
type AlicloudRedisInstanceSpec struct {
	// +optional
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf),message="IpRange is immutable."
	IpRange IpRangeRef `json:"ipRange"`

	// RedisTier defines service and capacity tier.
	// S = standard HA (master+replica, no read-only replica).
	// P = premium HA (master+replica + 1 read-only replica).
	// Both letter and number are mutable via ModifyInstanceSpec.
	// +kubebuilder:validation:Required
	RedisTier AlicloudRedisTier `json:"redisTier"`

	// EngineVersion is the Redis engine version. Immutable after creation.
	// +optional
	// +kubebuilder:default="5.0"
	// +kubebuilder:validation:Enum="5.0";"6.0";"7.0"
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf),message="engineVersion is immutable."
	EngineVersion string `json:"engineVersion"`

	// Parameters are passed to the AliCloud instance as runtime configuration.
	// +optional
	Parameters map[string]string `json:"parameters,omitempty"`

	// +optional
	AuthSecret *RedisAuthSecretSpec `json:"authSecret,omitempty"`
}

// AlicloudRedisInstanceStatus defines the observed state of AlicloudRedisInstance.
type AlicloudRedisInstanceStatus struct {
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

// AlicloudRedisInstance is the Schema for the alicloudredisinstances API
type AlicloudRedisInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AlicloudRedisInstanceSpec   `json:"spec,omitempty"`
	Status AlicloudRedisInstanceStatus `json:"status,omitempty"`
}

func (in *AlicloudRedisInstance) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

func (in *AlicloudRedisInstance) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *AlicloudRedisInstance) SpecificToFeature() featuretypes.FeatureName {
	return featuretypes.FeatureAlicloudRedisInstance
}

func (in *AlicloudRedisInstance) SpecificToProviders() []string {
	return []string{"alicloud"}
}

func (in *AlicloudRedisInstance) GetIpRangeRef() IpRangeRef {
	return in.Spec.IpRange
}

func (in *AlicloudRedisInstance) State() string {
	return in.Status.State
}

func (in *AlicloudRedisInstance) SetState(v string) {
	in.Status.State = v
}

func (in *AlicloudRedisInstance) CloneForPatchStatus() client.Object {
	return &AlicloudRedisInstance{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AlicloudRedisInstance",
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

// AlicloudRedisInstanceList contains a list of AlicloudRedisInstance
type AlicloudRedisInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AlicloudRedisInstance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AlicloudRedisInstance{}, &AlicloudRedisInstanceList{})
}
