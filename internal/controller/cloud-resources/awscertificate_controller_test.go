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

package cloudresources

import (
	"fmt"
	"strings"

	"github.com/kyma-project/cloud-manager/api"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	kcpscope "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	scopeprovider "github.com/kyma-project/cloud-manager/pkg/skr/common/scope/provider"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("AwsCertificate Controller", func() {

	It("Scenario: SKR AwsCertificate is created with valid Secret then deleted", func() {

		awsAccountLocal := infra.AwsMock().NewAccount()
		defer awsAccountLocal.Delete()

		scope := &cloudcontrolv1beta1.Scope{}
		scopeName := "aws-cert-test-scope"

		By("Given Scope exists", func() {
			kcpscope.Ignore.AddName(scopeName)
			Expect(CreateScopeAws(infra.Ctx(), infra, scope, awsAccountLocal.AccountId(), WithName(scopeName))).To(Succeed())
		})

		Expect(scope.Namespace).To(Equal(infra.KCP().Namespace()))
		Expect(scope.Name).To(Equal(scopeName))

		objName := "test-certificate"
		secretName := "test-cert-secret"
		infra.ScopeProvider().Add(scopeprovider.MatchingObj(objName, scope))

		awsMockLocal := awsAccountLocal.Region(scope.Spec.Region)

		By("And Given scope is ready", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), scope,
					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed())
		})

		secret := &corev1.Secret{}
		awsCertificate := &cloudresourcesv1beta1.AwsCertificate{}

		By("And Given Secret with certificate data exists", func() {
			Eventually(CreateCertificateSecret).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), secret,
					WithName(secretName),
					WithNamespace(infra.SKR().Namespace()),
				).
				Should(Succeed())
		})

		By("When AwsCertificate is created", func() {
			awsCertificate.Spec = cloudresourcesv1beta1.AwsCertificateSpec{
				SecretRef: klog.ObjectRef{
					Name:      secretName,
					Namespace: infra.SKR().Namespace(),
				},
			}

			Eventually(CreateAwsCertificate).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), awsCertificate,
					WithName(objName),
				).
				Should(Succeed())
		})

		By("Then AwsCertificate has finalizer and Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), awsCertificate,
					NewObjActions(),
					HaveFinalizer(api.CommonFinalizerDeletionHook),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

		By("And Then AwsCertificate status has ARN populated", func() {
			Expect(awsCertificate.Status.Arn).NotTo(BeEmpty(), "expected Status.Arn to be set")
			Expect(awsCertificate.Status.Arn).To(ContainSubstring("arn:aws:acm:"), "expected ARN to have correct format")
			Expect(awsCertificate.Status.Arn).To(ContainSubstring(":certificate/"), "expected ARN to reference certificate")
		})

		By("And Then AwsCertificate status has ExpirationDate populated", func() {
			Expect(awsCertificate.Status.ExpirationDate).NotTo(BeNil(), "expected ExpirationDate to be set")
		})

		By("And Then certificate exists in AWS ACM", func() {
			certDetail := awsMockLocal.GetCertificateByArn(awsCertificate.Status.Arn)
			Expect(certDetail).NotTo(BeNil(), "expected certificate to exist in ACM")
			Expect(*certDetail.CertificateArn).To(Equal(awsCertificate.Status.Arn))
		})

		By("And Then certificate has correct tags", func() {
			tags := awsMockLocal.GetCertificateTags(awsCertificate.Status.Arn)
			Expect(tags).NotTo(BeNil(), "expected certificate to have tags")

			// Convert tags to map for easier verification
			tagMap := make(map[string]string)
			for _, tag := range tags {
				if tag.Key != nil && tag.Value != nil {
					tagMap[*tag.Key] = *tag.Value
				}
			}

			// Verify cloud-manager name tag
			Expect(tagMap).To(HaveKeyWithValue(common.TagCloudManagerName, objName))

			// Verify ManagedBy tag
			Expect(tagMap).To(HaveKeyWithValue("kyma-project.io/managed-by", "cloud-manager"))

			// Verify Scope tag
			Expect(tagMap).To(HaveKeyWithValue("cloud-manager.kyma-project.io/scope", scopeName))

			// Verify Shoot tag
			Expect(tagMap).To(HaveKeyWithValue("cloud-manager.kyma-project.io/shoot", scope.Spec.ShootName))
		})

		originalArn := awsCertificate.Status.Arn

		By("When Secret data is updated", func() {
			// Load the secret fresh
			err := infra.SKR().Client().Get(infra.Ctx(), client.ObjectKeyFromObject(secret), secret)
			Expect(err).NotTo(HaveOccurred())

			// Generate a new valid certificate with different serial number and CA chain
			certPEM, keyPEM, caPEM, err := GenerateTestCertificate("updated.example.com", "Updated Test Org")
			Expect(err).NotTo(HaveOccurred())

			// Update the certificate data with new valid certificate
			secret.Data["tls.crt"] = certPEM
			secret.Data["tls.key"] = keyPEM
			secret.Data["ca.crt"] = caPEM

			// Save the updated secret
			err = infra.SKR().Client().Update(infra.Ctx(), secret)
			Expect(err).To(Succeed())
		})

		By("Then certificate is reimported with new data", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), awsCertificate,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
				).
				Should(Succeed())

			// Verify ARN didn't change (updated in place)
			Expect(awsCertificate.Status.Arn).To(Equal(originalArn))
		})

		By("When AwsCertificate is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsCertificate).
				Should(Succeed())
		})

		By("Then AwsCertificate is removed from K8s", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsCertificate).
				Should(Succeed())
		})

		By("And Then certificate is deleted from AWS ACM", func() {
			certDetail := awsMockLocal.GetCertificateByArn(awsCertificate.Status.Arn)
			Expect(certDetail).To(BeNil(), "expected certificate to be deleted from ACM")
		})
	})

	It("Scenario: AwsCertificate fails when Secret is missing", func() {

		awsAccountLocal := infra.AwsMock().NewAccount()
		defer awsAccountLocal.Delete()

		scope := &cloudcontrolv1beta1.Scope{}
		scopeName := "aws-cert-missing-secret-scope"

		By("Given Scope exists", func() {
			kcpscope.Ignore.AddName(scopeName)
			Expect(CreateScopeAws(infra.Ctx(), infra, scope, awsAccountLocal.AccountId(), WithName(scopeName))).To(Succeed())
		})

		objName := "test-certificate-no-secret"
		infra.ScopeProvider().Add(scopeprovider.MatchingObj(objName, scope))

		By("And Given scope is ready", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), scope,
					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed())
		})

		awsCertificate := &cloudresourcesv1beta1.AwsCertificate{}

		By("When AwsCertificate is created without existing Secret", func() {
			awsCertificate.Spec = cloudresourcesv1beta1.AwsCertificateSpec{
				SecretRef: klog.ObjectRef{
					Name:      "nonexistent-secret",
					Namespace: infra.SKR().Namespace(),
				},
			}

			Eventually(CreateAwsCertificate).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), awsCertificate,
					WithName(objName),
				).
				Should(Succeed())
		})

		By("Then AwsCertificate has Error condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), awsCertificate,
					NewObjActions(),
					NotHavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

		By("And Then status has no ARN", func() {
			Expect(awsCertificate.Status.Arn).To(BeEmpty(), "expected Status.Arn to be empty")
		})

		By("When AwsCertificate is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsCertificate).
				Should(Succeed())
		})
	})

	It("Scenario: AwsCertificate fails when Secret has invalid data", func() {

		awsAccountLocal := infra.AwsMock().NewAccount()
		defer awsAccountLocal.Delete()

		scope := &cloudcontrolv1beta1.Scope{}
		scopeName := "aws-cert-invalid-secret-scope"

		By("Given Scope exists", func() {
			kcpscope.Ignore.AddName(scopeName)
			Expect(CreateScopeAws(infra.Ctx(), infra, scope, awsAccountLocal.AccountId(), WithName(scopeName))).To(Succeed())
		})

		objName := "test-certificate-invalid-secret"
		secretName := "test-invalid-cert-secret"
		infra.ScopeProvider().Add(scopeprovider.MatchingObj(objName, scope))

		By("And Given scope is ready", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), scope,
					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed())
		})

		secret := &corev1.Secret{}
		awsCertificate := &cloudresourcesv1beta1.AwsCertificate{}

		By("And Given Secret with missing tls.crt key exists", func() {
			secret.SetName(secretName)
			secret.SetNamespace(infra.SKR().Namespace())
			secret.Type = corev1.SecretTypeOpaque // Use Opaque instead of TLS to bypass validation
			// Only provide tls.key, missing tls.crt
			secret.Data = map[string][]byte{
				"tls.key": []byte("-----BEGIN PRIVATE KEY-----\nKEY_DATA\n-----END PRIVATE KEY-----"),
			}
			Expect(infra.SKR().Client().Create(infra.Ctx(), secret)).To(Succeed())
		})

		By("When AwsCertificate is created", func() {
			awsCertificate.Spec = cloudresourcesv1beta1.AwsCertificateSpec{
				SecretRef: klog.ObjectRef{
					Name:      secretName,
					Namespace: infra.SKR().Namespace(),
				},
			}

			Eventually(CreateAwsCertificate).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), awsCertificate,
					WithName(objName),
				).
				Should(Succeed())
		})

		By("Then AwsCertificate has Error condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), awsCertificate,
					NewObjActions(),
					NotHavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

		By("And Then error message mentions missing tls.crt", func() {
			Eventually(func() error {
				// Reload object to get latest status
				key := client.ObjectKey{
					Name:      awsCertificate.Name,
					Namespace: awsCertificate.Namespace,
				}
				err := infra.SKR().Client().Get(infra.Ctx(), key, awsCertificate)
				if err != nil {
					return err
				}

				// Check condition exists
				cond := meta.FindStatusCondition(awsCertificate.Status.Conditions,
					cloudresourcesv1beta1.ConditionTypeReady)
				if cond == nil {
					return fmt.Errorf("Ready condition not found")
				}

				// Verify condition properties
				if cond.Status != metav1.ConditionFalse {
					return fmt.Errorf("condition status is %s, expected False", cond.Status)
				}
				if !strings.Contains(cond.Message, "tls.crt") {
					return fmt.Errorf("message doesn't contain 'tls.crt': %s", cond.Message)
				}

				return nil
			}).Should(Succeed())
		})

		By("When AwsCertificate is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsCertificate).
				Should(Succeed())
		})
	})

	It("Scenario: AwsCertificate deletion is blocked when certificate is in use", func() {

		awsAccountLocal := infra.AwsMock().NewAccount()
		defer awsAccountLocal.Delete()

		scope := &cloudcontrolv1beta1.Scope{}
		scopeName := "aws-cert-in-use-scope"

		By("Given Scope exists", func() {
			kcpscope.Ignore.AddName(scopeName)
			Expect(CreateScopeAws(infra.Ctx(), infra, scope, awsAccountLocal.AccountId(), WithName(scopeName))).To(Succeed())
		})

		objName := "test-certificate-in-use"
		secretName := "test-cert-in-use-secret"
		infra.ScopeProvider().Add(scopeprovider.MatchingObj(objName, scope))

		awsMockLocal := awsAccountLocal.Region(scope.Spec.Region)

		By("And Given scope is ready", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), scope,
					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed())
		})

		secret := &corev1.Secret{}
		awsCertificate := &cloudresourcesv1beta1.AwsCertificate{}

		By("And Given Secret with certificate data exists", func() {
			Eventually(CreateCertificateSecret).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), secret,
					WithName(secretName),
					WithNamespace(infra.SKR().Namespace()),
				).
				Should(Succeed())
		})

		By("And Given AwsCertificate is created and Ready", func() {
			awsCertificate.Spec = cloudresourcesv1beta1.AwsCertificateSpec{
				SecretRef: klog.ObjectRef{
					Name:      secretName,
					Namespace: infra.SKR().Namespace(),
				},
			}

			Eventually(CreateAwsCertificate).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), awsCertificate,
					WithName(objName),
				).
				Should(Succeed())

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), awsCertificate,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

		By("And Given certificate is marked as in-use in AWS", func() {
			awsMockLocal.SetCertificateInUse(awsCertificate.Status.Arn, true)
		})

		By("When AwsCertificate is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsCertificate).
				Should(Succeed())
		})

		By("Then AwsCertificate has Error condition about being in use", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), awsCertificate,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeDeleteWhileUsed),
				).
				Should(Succeed())

			cond := meta.FindStatusCondition(awsCertificate.Status.Conditions, cloudresourcesv1beta1.ConditionTypeDeleteWhileUsed)
			Expect(cond).NotTo(BeNil())
			Expect(cond.Status).To(Equal(metav1.ConditionTrue))
			Expect(cond.Message).To(ContainSubstring("in use"))
		})

		By("And Then AwsCertificate still has finalizer (not removed)", func() {
			Expect(awsCertificate.Finalizers).To(ContainElement(api.CommonFinalizerDeletionHook))
		})

		By("When certificate is marked as not in-use", func() {
			awsMockLocal.SetCertificateInUse(awsCertificate.Status.Arn, false)
		})

		By("Then AwsCertificate is eventually removed from K8s", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsCertificate).
				Should(Succeed(), "certificate should be deleted after no longer in use")
		})

		By("And Then certificate is deleted from AWS ACM", func() {
			certDetail := awsMockLocal.GetCertificateByArn(awsCertificate.Status.Arn)
			Expect(certDetail).To(BeNil(), "expected certificate to be deleted from ACM")
		})
	})
})
