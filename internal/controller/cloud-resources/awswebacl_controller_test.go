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
	"github.com/kyma-project/cloud-manager/api"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	kcpscope "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var _ = Describe("AwsWebAcl Controller", Focus, func() {
	It("Scenario: SKR AwsWebAcl is created then deleted", func() {
		kymaName := infra.SkrKymaRef().Name

		awsAccountLocal := infra.AwsMock().NewAccount()
		defer awsAccountLocal.Delete()

		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope is created", func() {
			// Tell Scope reconciler to ignore this kymaName
			kcpscope.Ignore.AddName(kymaName)

			Eventually(CreateScopeAws).
				WithArguments(infra.Ctx(), infra, scope, awsAccountLocal.AccountId(), WithName(kymaName)).
				Should(Succeed())
		})
		awsWebAcl := &cloudresourcesv1beta1.AwsWebAcl{}

		By("Given scope exists", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), scope, NewObjActions()).
				Should(Succeed())
		})

		By("When AwsWebAcl is created", func() {
			Eventually(CreateAwsWebAcl).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), awsWebAcl,
					WithName("67ba9c61-8826-4349-873c-d30e9756aaec"),
				).Should(Succeed())
		})

		By("Then AwsWebAcl is loaded", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), awsWebAcl,
					NewObjActions(),
					HaveFinalizer(api.CommonFinalizerDeletionHook),
				).
				Should(Succeed(), "expected AwsWebAcl to have finalizer")
		})

		By("And Then AwsWebAcl has finalizer", func() {
			Expect(controllerutil.ContainsFinalizer(awsWebAcl, api.CommonFinalizerDeletionHook)).
				To(BeTrue(), "expected AwsWebAcl to have finalizer, but it does not")
		})

		By("When AwsWebAcl is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsWebAcl).
				Should(Succeed())
		})

		By("Then AwsWebAcl is deleted", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsWebAcl).
				Should(Succeed())
		})
	})
})
