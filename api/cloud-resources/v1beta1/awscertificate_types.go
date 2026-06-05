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
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	featuretypes "github.com/kyma-project/cloud-manager/pkg/feature/types"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

// AwsCertificateSpec defines the desired state of AwsCertificate
type AwsCertificateSpec struct {
	// SecretRef references a Secret containing certificate data
	// Required keys: "tls.crt", "tls.key"
	// Optional keys: "ca.crt" (certificate chain)
	// +kubebuilder:validation:Required
	SecretRef klog.ObjectRef `json:"secretRef"`
}

// AwsCertificateStatus defines the observed state of AwsCertificate
type AwsCertificateStatus struct {
	// ARN of the imported certificate in ACM
	// +optional
	Arn string `json:"arn,omitempty"`

	// ExpirationDate of the certificate (from ACM)
	// +optional
	ExpirationDate *metav1.Time `json:"expirationDate,omitempty"`

	// List of status conditions to indicate the status of a AwsCertificate
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// +optional
	State string `json:"state,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={kyma-cloud-manager}
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"
// +kubebuilder:printcolumn:name="ARN",type="string",JSONPath=".status.arn"
// +kubebuilder:printcolumn:name="Expiration",type="date",JSONPath=".status.expirationDate"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// AwsCertificate is the Schema for the awscertificates API
type AwsCertificate struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// spec defines the desired state of AwsCertificate
	// +required
	Spec AwsCertificateSpec `json:"spec"`

	// status defines the observed state of AwsCertificate
	// +optional
	Status AwsCertificateStatus `json:"status,omitempty"`
}

func (in *AwsCertificate) ObservedGeneration() int64 {
	return in.Status.ObservedGeneration
}

func (in *AwsCertificate) SetObservedGeneration(i int64) {
	in.Status.ObservedGeneration = i
}

func (in *AwsCertificate) GetStatus() any {
	return &in.Status
}

func (in *AwsCertificate) SetStatusProviderError(msg string) {
	in.Status.State = ReasonProviderError
	meta.SetStatusCondition(&in.Status.Conditions, metav1.Condition{
		Type:               ConditionTypeReady,
		Status:             metav1.ConditionFalse,
		ObservedGeneration: in.Generation,
		Reason:             ReasonProviderError,
		Message:            msg,
	})
}

func (in *AwsCertificate) SetStatusReady() {
	in.Status.State = StateReady
	meta.SetStatusCondition(&in.Status.Conditions, metav1.Condition{
		Type:               ConditionTypeReady,
		Status:             metav1.ConditionTrue,
		ObservedGeneration: in.Generation,
		Reason:             ReasonReady,
		Message:            ReasonReady,
	})
}

func (in *AwsCertificate) SetStatusProcessing() {
	in.Status.State = StateProcessing
	meta.SetStatusCondition(&in.Status.Conditions, metav1.Condition{
		Type:               ConditionTypeReady,
		Status:             metav1.ConditionFalse,
		ObservedGeneration: in.Generation,
		Reason:             v1beta1.ReasonProcessing,
		Message:            v1beta1.ReasonProcessing,
	})
}

func (in *AwsCertificate) SetStatusDeleting() {
	in.Status.State = StateDeleting
	meta.SetStatusCondition(&in.Status.Conditions, metav1.Condition{
		Type:               ConditionTypeReady,
		Status:             metav1.ConditionUnknown,
		ObservedGeneration: in.Status.ObservedGeneration,
		Reason:             v1beta1.ReasonDeleting,
		Message:            v1beta1.ReasonDeleting,
	})
}
func (in *AwsCertificate) Conditions() *[]metav1.Condition { return &in.Status.Conditions }

func (in *AwsCertificate) GetObjectMeta() *metav1.ObjectMeta { return &in.ObjectMeta }

func (in *AwsCertificate) SpecificToFeature() featuretypes.FeatureName {
	return featuretypes.FeatureCertificate
}

func (in *AwsCertificate) SpecificToProviders() []string { return []string{"aws"} }

func (in *AwsCertificate) State() string {
	return in.Status.State
}

func (in *AwsCertificate) SetState(v string) {
	in.Status.State = v
}

// +kubebuilder:object:root=true

// AwsCertificateList contains a list of AwsCertificate
type AwsCertificateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AwsCertificate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AwsCertificate{}, &AwsCertificateList{})
}
