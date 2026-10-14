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
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	kcpscope "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	scopeprovider "github.com/kyma-project/cloud-manager/pkg/skr/common/scope/provider"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

var _ = Describe("Feature: SKR Certificate", func() {

	It("Scenario: SKR Certificate with chain-in-tls.crt is imported and deleted", func() {

		awsAccount := infra.AwsMock().NewAccount()
		defer awsAccount.Delete()

		scope := &cloudcontrolv1beta1.Scope{}
		scopeName := "cert-chain-test-scope"

		By("Given Scope exists", func() {
			kcpscope.Ignore.AddName(scopeName)
			Expect(CreateScopeAws(infra.Ctx(), infra, scope, awsAccount.AccountId(), WithName(scopeName))).To(Succeed())
		})

		certName := "cert-chain-test"
		infra.ScopeProvider().Add(scopeprovider.MatchingObj(certName, scope))

		By("And Given scope is ready", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), scope,
					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed())
		})

		// tls.crt holds leaf + intermediate concatenated (Gardener-style); only the leaf
		// is imported — intermediates in tls.crt are ignored, ca.crt carries the chain.
		tlsCrt, tlsKey, err := GenerateTestCertificateWithChain("test.example.com", "Test Org")
		Expect(err).NotTo(HaveOccurred())

		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "cert-chain-secret",
				Namespace: infra.SKR().Namespace(),
			},
			Type: corev1.SecretTypeTLS,
			Data: map[string][]byte{
				"tls.crt": tlsCrt,
				"tls.key": tlsKey,
			},
		}

		By("And Given TLS secret with chain-in-tls.crt exists", func() {
			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.SKR().Client(), secret).
				Should(Succeed())
		})

		cert := &cloudresourcesv1beta1.Certificate{}

		By("When Certificate is created referencing that secret", func() {
			cert.Name = certName
			cert.Spec = cloudresourcesv1beta1.CertificateSpec{
				SecretRef: klog.ObjectRef{
					Name:      secret.Name,
					Namespace: secret.Namespace,
				},
			}
			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.SKR().Client(), cert).
				Should(Succeed())
		})

		By("Then Certificate becomes Ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					cert,
					NewObjActions(),
					HavingState(cloudresourcesv1beta1.StateReady),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

		By("And Then Certificate has a ProviderId (ARN) set", func() {
			Expect(cert.Status.ProviderId).NotTo(BeEmpty())
		})

		By("When Certificate is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), cert).
				Should(Succeed())
		})

		By("Then Certificate is gone", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), cert).
				Should(Succeed())
		})

		By("And Then secret is cleaned up", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), secret).
				Should(Succeed())
		})
	})
})
