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

// +kubebuilder:validation:Enum=S1;S2;S3;S4;S5;S6;S7;P1;P2;P3;P4;P5;P6
type AwsRedisTier string

const (
	AwsRedisTierS1 AwsRedisTier = "S1"
	AwsRedisTierS2 AwsRedisTier = "S2"
	AwsRedisTierS3 AwsRedisTier = "S3"
	AwsRedisTierS4 AwsRedisTier = "S4"
	AwsRedisTierS5 AwsRedisTier = "S5"
	AwsRedisTierS6 AwsRedisTier = "S6"
	AwsRedisTierS7 AwsRedisTier = "S7"
	AwsRedisTierS8 AwsRedisTier = "S8"

	AwsRedisTierP1 AwsRedisTier = "P1"
	AwsRedisTierP2 AwsRedisTier = "P2"
	AwsRedisTierP3 AwsRedisTier = "P3"
	AwsRedisTierP4 AwsRedisTier = "P4"
	AwsRedisTierP5 AwsRedisTier = "P5"
	AwsRedisTierP6 AwsRedisTier = "P6"
)

// AwsRedisInstanceSpec defines the desired state of AwsRedisInstance
type AwsRedisInstanceSpec struct {
	// +optional
	IpRange IpRangeRef `json:"ipRange"`

	// +optional
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="AuthSecret is immutable."
	AuthSecret *RedisAuthSecretSpec `json:"authSecret,omitempty"`

	// Defines Service Tier and Capacity Tier. RedisTiers starting with 'S' are Standard service tier. RedisTiers starting with 'P' are premium servicetier. Number next to service tier represents capacity tier.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self.startsWith('S') && oldSelf.startsWith('S') || self.startsWith('P') && oldSelf.startsWith('P')), message="Service tier cannot be changed within redisTier. Only capacity tier can be changed."
	RedisTier AwsRedisTier `json:"redisTier"`

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
