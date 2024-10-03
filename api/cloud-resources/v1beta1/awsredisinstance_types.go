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

// AwsRedisInstanceSpec defines the desired state of AwsRedisInstance
// +kubebuilder:validation:XValidation:rule=(self.authEnabled == false || self.transitEncryptionEnabled == true), message="authEnabled can only be true if TransitEncryptionEnabled is also true"
type AwsRedisInstanceSpec struct {
	// +optional
	IpRange IpRangeRef `json:"ipRange"`

	// +optional
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="AuthSecret is immutable."
	AuthSecret *AuthSecretSpec `json:"authSecret,omitempty"`

	// +kubebuilder:validation:Required
	CacheNodeType string `json:"cacheNodeType"`

	// +optional
	// +kubebuilder:default="7.0"
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="EngineVersion is immutable."
	EngineVersion string `json:"engineVersion"`

	// +optional
	// +kubebuilder:default=false
	AutoMinorVersionUpgrade bool `json:"autoMinorVersionUpgrade"`

	// +optional
	// +kubebuilder:default=false
	TransitEncryptionEnabled bool `json:"transitEncryptionEnabled"`

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
// +kubebuilder:resource:categories={kyma-cloud-manager}
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

func (in *AwsRedisInstance) CloneForPatchStatus() client.Object {
	return &AwsRedisInstance{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AwsRedisInstance",
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

// AwsRedisInstanceList contains a list of AwsRedisInstance
type AwsRedisInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AwsRedisInstance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AwsRedisInstance{}, &AwsRedisInstanceList{})
}
