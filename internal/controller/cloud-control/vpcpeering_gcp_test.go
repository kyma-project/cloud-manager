package cloudcontrol

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	scopePkg "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Feature: KCP VpcPeering", func() {
	It("Scenario: KCP GCP VpcPeering is created", func() {
		const (
			kymaName           = "57bc9639-d752-4f67-8b9e-7cd12514575f"
			peeringName        = "peering-sap-gcp-skr-dev-cust-00002-to-sap-sc-learn"
			remoteVpc          = "default"
			remoteProject      = "sap-sc-learn"
			remoteRefNamespace = "kcp-system"
			remoteRefName      = "skr-gcp-vpcpeering"
			importCustomRoutes = true
		)

		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			scopePkg.Ignore.AddName(kymaName)

			Eventually(CreateScopeGcp).
				WithArguments(infra.Ctx(), infra, scope, WithName(kymaName)).
				Should(Succeed())
		})

		vpcpeering := &cloudcontrolv1beta1.VpcPeering{}

		By("When KCP VpcPeering is created", func() {
			Eventually(CreateKcpVpcPeering).
				WithArguments(infra.Ctx(), infra.KCP().Client(), vpcpeering,
					WithName(peeringName),
					WithKcpVpcPeeringRemoteRef(remoteRefNamespace, remoteRefName),
					WithKcpVpcPeeringSpecScope(kymaName),
					WithKcpVpcPeeringSpecGCP(remoteVpc, remoteProject, peeringName, importCustomRoutes),
				).
				Should(Succeed())
		})

		By("Then KCP VpcPeering has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), vpcpeering,
					NewObjActions(),
					HaveFinalizer(cloudcontrolv1beta1.FinalizerName),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

	})
})
