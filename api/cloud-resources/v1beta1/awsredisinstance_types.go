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

// AwsRedisInstanceSpec defines the desired state of AwsRedisInstance
type AwsRedisInstanceSpec struct {
	// +optional
	IpRange IpRangeRef `json:"ipRange"`

	// +optional
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="AuthSecret is immutable."
	AuthSecret *AuthSecretSpec `json:"authSecret,omitempty"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="CacheNodeType is immutable."
	CacheNodeType string `json:"cacheNodeType"`

	// +optional
	// +kubebuilder:default="7.0"
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="EngineVersion is immutable."
	EngineVersion string `json:"engineVersion"`

	// +optional
	// +kubebuilder:default=false
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="AutoMinorVersionUpgrade is immutable."
	AutoMinorVersionUpgrade bool `json:"autoMinorVersionUpgrade"`

	// +optional
	Parameters map[string]string `json:"parameters,omitempty"`
}

// AwsRedisInstanceStatus defines the observed state of AwsRedisInstance
type AwsRedisInstanceStatus struct {
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
// AwsRedisInstance is the Schema for the awsredisinstances API
type AwsRedisInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AwsRedisInstanceSpec   `json:"spec,omitempty"`
	Status AwsRedisInstanceStatus `json:"status,omitempty"`
}

func (in *AwsRedisInstance) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

func (in *AwsRedisInstance) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *AwsRedisInstance) SpecificToFeature() featuretypes.FeatureName {
	return featuretypes.FeatureRedis
}

func (in *AwsRedisInstance) SpecificToProviders() []string {
	return []string{"aws"}
}

func (in *AwsRedisInstance) GetIpRangeRef() IpRangeRef {
	return in.Spec.IpRange
}

func (in *AwsRedisInstance) State() string {
	return in.Status.State
}

func (in *AwsRedisInstance) SetState(v string) {
	in.Status.State = v
}

//+kubebuilder:object:root=true

// AwsRedisInstanceList contains a list of AwsRedisInstance
type AwsRedisInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AwsRedisInstance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AwsRedisInstance{}, &AwsRedisInstanceList{})
}
