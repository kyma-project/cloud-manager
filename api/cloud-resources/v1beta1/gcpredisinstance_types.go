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

// +kubebuilder:validation:Enum=S1;S2;S3;S4;S5;S6;S7;S8;P1;P2;P3;P4;P5;P6;P7
type GcpRedisTier string

const (
	GcpRedisTierS1 GcpRedisTier = "S1"
	GcpRedisTierS2 GcpRedisTier = "S2"
	GcpRedisTierS3 GcpRedisTier = "S3"
	GcpRedisTierS4 GcpRedisTier = "S4"
	GcpRedisTierS5 GcpRedisTier = "S5"
	GcpRedisTierS6 GcpRedisTier = "S6"
	GcpRedisTierS7 GcpRedisTier = "S7"
	GcpRedisTierS8 GcpRedisTier = "S8"

	GcpRedisTierP1 GcpRedisTier = "P1"
	GcpRedisTierP2 GcpRedisTier = "P2"
	GcpRedisTierP3 GcpRedisTier = "P3"
	GcpRedisTierP4 GcpRedisTier = "P4"
	GcpRedisTierP5 GcpRedisTier = "P5"
	GcpRedisTierP6 GcpRedisTier = "P6"
)

type TimeOfDay struct {
	// Hours of day in 24 hour format. Should be from 0 to 23.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=23
	Hours int32 `json:"hours"`

	// Minutes of hour of day. Must be from 0 to 59.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=59
	Minutes int32 `json:"minutes"`
}

type DayOfWeekPolicy struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=MONDAY;TUESDAY;WEDNESDAY;THURSDAY;FRIDAY;SATURDAY;SUNDAY;
	Day string `json:"day"`

	// Start time of the window in UTC time.
	// +kubebuilder:validation:Required
	StartTime TimeOfDay `json:"startTime"`
}

// +kubebuilder:validation:MinProperties=1
// +kubebuilder:validation:MaxProperties=1
type MaintenancePolicy struct {
	DayOfWeek *DayOfWeekPolicy `json:"dayOfWeek,omitempty"`
}

// GcpRedisInstanceSpec defines the desired state of GcpRedisInstance
type GcpRedisInstanceSpec struct {

	// +optional
	IpRange IpRangeRef `json:"ipRange"`

	// Defines Service Tier and Capacity Tier. RedisTiers starting with 'S' are Standard service tier. RedisTiers starting with 'P' are premium servicetier. Number next to service tier represents capacity tier.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self.startsWith('S') && oldSelf.startsWith('S') || self.startsWith('P') && oldSelf.startsWith('P')), message="Service tier cannot be changed within redisTier. Only capacity tier can be changed."
	RedisTier GcpRedisTier `json:"redisTier"`

	// The version of Redis software.
	// +optional
	// +kubebuilder:default="REDIS_7_0"
	// +kubebuilder:validation:Enum=REDIS_7_2;REDIS_7_0;REDIS_6_X
	// +kubebuilder:validation:XValidation:rule=(self != "REDIS_7_0" || oldSelf == "REDIS_7_0" || oldSelf == "REDIS_6_X"), message="redisVersion cannot be downgraded."
	// +kubebuilder:validation:XValidation:rule=(self != "REDIS_7_2" || oldSelf == "REDIS_7_2" || oldSelf == "REDIS_7_0" || oldSelf == "REDIS_6_X"), message="redisVersion cannot be downgraded."
	// +kubebuilder:validation:XValidation:rule=(self != "REDIS_6_X" || oldSelf == "REDIS_6_X"), message="redisVersion cannot be downgraded."
	RedisVersion string `json:"redisVersion"`

	// Indicates whether OSS Redis AUTH is enabled for the instance.
	// +optional
	// +kubebuilder:default=false
	AuthEnabled bool `json:"authEnabled"`

	// Redis configuration parameters, according to http://redis.io/topics/config.
	// See docs for the list of the supported parameters
	// +optional
	RedisConfigs map[string]string `json:"redisConfigs"`

	// +optional
	AuthSecret *RedisAuthSecretSpec `json:"authSecret,omitempty"`

	// The maintenance policy for the instance.
	// If not provided, maintenance events can be performed at any time.
	// +optional
	MaintenancePolicy *MaintenancePolicy `json:"maintenancePolicy,omitempty"`
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
// +kubebuilder:resource:categories={kyma-cloud-manager}
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

func (in *GcpRedisInstance) CloneForPatchStatus() client.Object {
	return &GcpRedisInstance{
		TypeMeta: metav1.TypeMeta{
			Kind:       "GcpRedisInstance",
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

// GcpRedisInstanceList contains a list of GcpRedisInstance
type GcpRedisInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GcpRedisInstance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GcpRedisInstance{}, &GcpRedisInstanceList{})
}
