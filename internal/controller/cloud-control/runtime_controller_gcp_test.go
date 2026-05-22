package cloudcontrol

import (
	"context"
	"fmt"

	"cloud.google.com/go/securitycentermanagement/apiv1/securitycentermanagementpb"
	"github.com/kyma-project/cloud-manager/api"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	gcpmock2 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/mock2"
	kcpruntime "github.com/kyma-project/cloud-manager/pkg/kcp/runtime"
	kcpsubscription "github.com/kyma-project/cloud-manager/pkg/kcp/subscription"
	kcpvpcnetwork "github.com/kyma-project/cloud-manager/pkg/kcp/vpcnetwork"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	"github.com/kyma-project/cloud-manager/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
)

var _ = Describe("Feature: Runtime", func() {

	gcpSccServiceHasState := func(ctx context.Context, mock gcpmock2.Store, serviceId string, state securitycentermanagementpb.SecurityCenterService_EnablementState) error {
		name := fmt.Sprintf("projects/%s/locations/global/securityCenterServices/%s", mock.ProjectId(), serviceId)
		svc, err := mock.GetSecurityCenterService(ctx, name)
		if err != nil {
			return err
		}
		if svc.IntendedEnablementState != state {
			return fmt.Errorf("SCC service %q: expected state %v, got %v", serviceId, state, svc.IntendedEnablementState)
		}
		return nil
	}

	It("Scenario: GCP Runtime with Gardener network is created and deleted", func() {

		name := "be84afc9-a96c-4586-b60e-4f5e1b8e976f"
		shootName := "t-4f5e1b8e976f"
		secretBindingName := "secret-binding-4f5e1b8e976f"

		var runtime *infrastructuremanagerv1.Runtime
		subscription := &cloudcontrolv1beta1.Subscription{}
		vpcNetwork := &cloudcontrolv1beta1.VpcNetwork{}

		gardenerVpcName := common.GardenerVpcName(infra.GardenerNamespace(), shootName)

		gcpVpcNetworkId := "vpc-4f5e1b8e976f"
		gcpRouterId := "router-4f5e1b8e976f"

		kcpsubscription.Ignore.AddName(secretBindingName)
		kcpvpcnetwork.Ignore.AddName(name)

		gcpMock := infra.GcpMock2().NewSubscription("rt")
		defer gcpMock.Delete()

		By("When Runtime is created", func() {
			runtime = infrastructuremanagerv1.NewSimpleRuntimeBuilder().
				WithName(name).
				WithProvider(cloudcontrolv1beta1.ProviderGCP).
				WithShootName(shootName).
				WithBindingName(secretBindingName).
				Build()

			Expect(CreateObj(infra.Ctx(), infra.KCP().Client(), runtime)).
				To(Succeed())

			_, err := composed.PatchObjAddFinalizer(infra.Ctx(), api.CommonFinalizerDeletionHook, runtime, infra.KCP().Client())
			Expect(err).NotTo(HaveOccurred())
		})

		By("Then Subscription is created with runtime's secret binding", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), subscription, NewObjActions(WithName(secretBindingName))).
				Should(Succeed())
		})

		By("And Then Subscription has labels as Runtime", func() {
			for _, labelName := range cloudcontrolv1beta1.ScopeLabels {
				rVal, ok := runtime.Labels[labelName]
				Expect(ok).To(BeTrue(), "unexpected logical error - runtime should have label %s set", labelName)
				Expect(subscription.Labels).To(HaveKeyWithValue(labelName, rVal), "subscription should have label %s", labelName)
			}
		})

		By("And Then Subscription has label managed-by cloud-manager", func() {
			Expect(subscription.Labels).To(HaveKeyWithValue(util.WellKnownK8sLabelManagedBy, util.DefaultCloudManagerManagedByLabelValue))
		})

		By("And Then VpcNetwork is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), vpcNetwork, NewObjActions(WithName(name))).
				Should(Succeed())
		})

		By("And Then VpcNetwork type is Gardener", func() {
			Expect(vpcNetwork.Spec.Type).To(Equal(cloudcontrolv1beta1.VpcNetworkTypeGardener))
		})

		By("And Then VpcNetwork is in the Subscription", func() {
			Expect(vpcNetwork.Spec.Subscription).To(Equal(subscription.Name))
		})

		By("And Then VpcNetwork name matches Gardener naming", func() {
			Expect(vpcNetwork.Spec.VpcNetworkName).To(HaveValue(Equal(gardenerVpcName)))
		})

		By("And Then VpcNetwork has labels as Runtime", func() {
			for _, labelName := range cloudcontrolv1beta1.ScopeLabels {
				rVal, ok := runtime.Labels[labelName]
				Expect(ok).To(BeTrue(), "unexpected logical error - runtime should have label %s set", labelName)
				Expect(vpcNetwork.Labels).To(HaveKeyWithValue(labelName, rVal), "vpcNetwork should have label %s", labelName)
			}
		})

		By("And Then VpcNetwork has label managed-by cloud-manager", func() {
			Expect(vpcNetwork.Labels).To(HaveKeyWithValue(util.WellKnownK8sLabelManagedBy, util.DefaultCloudManagerManagedByLabelValue))
		})

		By("When Subscription is ready", func() {
			err := composed.NewStatusPatcher(subscription).
				MutateStatus(func(sub *cloudcontrolv1beta1.Subscription) {
					sub.SetStatusReady()
					sub.Status.Provider = cloudcontrolv1beta1.ProviderGCP
					sub.Status.SubscriptionInfo = &cloudcontrolv1beta1.SubscriptionInfo{
						Gcp: &cloudcontrolv1beta1.SubscriptionInfoGcp{
							Project: gcpMock.ProjectId(),
						},
					}
				}).
				Patch(infra.Ctx(), infra.KCP().Client())
			Expect(err).NotTo(HaveOccurred())
		})

		By("And When VpcNetwork is ready", func() {
			err := composed.NewStatusPatcher(vpcNetwork).
				MutateStatus(func(net *cloudcontrolv1beta1.VpcNetwork) {
					net.SetStatusProvisioned()
					net.Status.Identifiers.Name = ptr.Deref(vpcNetwork.Spec.VpcNetworkName, "")
					net.Status.Identifiers.Vpc = gcpVpcNetworkId
					net.Status.Identifiers.Router = gcpRouterId
				}).
				Patch(infra.Ctx(), infra.KCP().Client())
			Expect(err).NotTo(HaveOccurred())
		})

		By("Then Runtime is successfully reconciled", func() {
			// Runtime is not owned by CloudManager, and it does not write its conditions
			// so there's no way to observe the reconciliation progress other than with Tracker
			Eventually(kcpruntime.Tracker.IsReconciledWith).
				WithArguments(runtime.Name, composed.ReconciliationLabelSuccess).
				Should(Succeed())
		})

		if feature.RuntimeSecurityGcp.Value(infra.Ctx()) {

			By("Then Runtime is annotated as security handled", func() {
				Eventually(LoadAndCheck).
					WithArguments(infra.Ctx(), infra.KCP().Client(), runtime, NewObjActions(), HavingAnnotation(cloudcontrolv1beta1.RuntimeSecurityStatusAnnotation, "Ready")).
					Should(Succeed())
			})

			By("When Runtime security is enabled", func() {
				Expect(LoadAndCheck(infra.Ctx(), infra.KCP().Client(), runtime, NewObjActions())).
					To(Succeed())

				// reset the tracker so next reconciliation can be tracked
				kcpruntime.Tracker.Clear(runtime.Name)

				_, err := composed.PatchObjMergeLabel(infra.Ctx(), runtime, infra.KCP().Client(), common.TmpRuntimeSecurityEnabledLabel, "true")
				Expect(err).NotTo(HaveOccurred())
			})

			By("Then Runtime is successfully reconciled", func() {
				Eventually(kcpruntime.Tracker.IsReconciledWith).
					WithArguments(runtime.Name, composed.ReconciliationLabelSuccess).
					Should(Succeed())
			})

			By("Then GCP SCC service security-health-analytics is ENABLED", func() {
				Eventually(gcpSccServiceHasState).
					WithArguments(infra.Ctx(), gcpMock, "security-health-analytics", securitycentermanagementpb.SecurityCenterService_ENABLED).
					Should(Succeed())
			})

			By("Then GCP SCC service web-security-scanner is ENABLED", func() {
				Eventually(gcpSccServiceHasState).
					WithArguments(infra.Ctx(), gcpMock, "web-security-scanner", securitycentermanagementpb.SecurityCenterService_ENABLED).
					Should(Succeed())
			})

			By("Then GCP SCC service event-threat-detection is ENABLED", func() {
				Eventually(gcpSccServiceHasState).
					WithArguments(infra.Ctx(), gcpMock, "event-threat-detection", securitycentermanagementpb.SecurityCenterService_ENABLED).
					Should(Succeed())
			})

			By("Then GCP SCC service vm-threat-detection is ENABLED", func() {
				Eventually(gcpSccServiceHasState).
					WithArguments(infra.Ctx(), gcpMock, "vm-threat-detection", securitycentermanagementpb.SecurityCenterService_ENABLED).
					Should(Succeed())
			})

			By("Then GCP SCC service GCE_VULNERABILITY_ASSESSMENT is ENABLED", func() {
				Eventually(gcpSccServiceHasState).
					WithArguments(infra.Ctx(), gcpMock, "GCE_VULNERABILITY_ASSESSMENT", securitycentermanagementpb.SecurityCenterService_ENABLED).
					Should(Succeed())
			})

		} // if security feature enabled for GCP

		// DELETE ===============================================================

		By("When Runtime is deleted", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), runtime)).To(Succeed())
		})

		if feature.RuntimeSecurityGcp.Value(infra.Ctx()) {

			By("Then GCP SCC service security-health-analytics is DISABLED", func() {
				Eventually(gcpSccServiceHasState).
					WithArguments(infra.Ctx(), gcpMock, "security-health-analytics", securitycentermanagementpb.SecurityCenterService_DISABLED).
					Should(Succeed())
			})

			By("Then GCP SCC service web-security-scanner is DISABLED", func() {
				Eventually(gcpSccServiceHasState).
					WithArguments(infra.Ctx(), gcpMock, "web-security-scanner", securitycentermanagementpb.SecurityCenterService_DISABLED).
					Should(Succeed())
			})

			By("Then GCP SCC service event-threat-detection is DISABLED", func() {
				Eventually(gcpSccServiceHasState).
					WithArguments(infra.Ctx(), gcpMock, "event-threat-detection", securitycentermanagementpb.SecurityCenterService_DISABLED).
					Should(Succeed())
			})

			By("Then GCP SCC service vm-threat-detection is DISABLED", func() {
				Eventually(gcpSccServiceHasState).
					WithArguments(infra.Ctx(), gcpMock, "vm-threat-detection", securitycentermanagementpb.SecurityCenterService_DISABLED).
					Should(Succeed())
			})

			By("Then GCP SCC service GCE_VULNERABILITY_ASSESSMENT is DISABLED", func() {
				Eventually(gcpSccServiceHasState).
					WithArguments(infra.Ctx(), gcpMock, "GCE_VULNERABILITY_ASSESSMENT", securitycentermanagementpb.SecurityCenterService_DISABLED).
					Should(Succeed())
			})

		} // if security feature enabled for GCP

		By("Then VpcNetwork is deleted", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), vpcNetwork).
				Should(Succeed())
		})

		By("And Then Subscription is not deleted", func() {
			Expect(LoadAndCheck(infra.Ctx(), infra.KCP().Client(), subscription, NewObjActions())).
				To(Succeed())
			Expect(subscription.DeletionTimestamp).To(BeNil(), "subscription should not be marked for deletion")
		})

		By("// cleanup: delete Subscription", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), subscription)).
				To(Succeed())
			Expect(IsDeleted(infra.Ctx(), infra.KCP().Client(), subscription)).
				To(Succeed())
		})

		By("// cleanup: delete Runtime", func() {
			_, err := composed.PatchObjRemoveFinalizer(infra.Ctx(), api.CommonFinalizerDeletionHook, runtime, infra.KCP().Client())
			Expect(err).NotTo(HaveOccurred())
			Expect(IsDeleted(infra.Ctx(), infra.KCP().Client(), runtime)).
				To(Succeed())
		})
	})

	It("Scenario: GCP Runtime with Kyma network is created and deleted", func() {
		name := "898299b5-7ba5-407a-9f7a-eb38cf14e1be"
		shootName := "t-eb38cf14e1be"
		secretBindingName := "secret-binding-eb38cf14e1be"
		vpcNetworkName := "5aba3b50-edde-4310-a697-236af3cd4b0d"

		var runtime *infrastructuremanagerv1.Runtime
		subscription := &cloudcontrolv1beta1.Subscription{}
		vpcNetwork := &cloudcontrolv1beta1.VpcNetwork{}

		gcpVpcNetworkId := "vpc-eb38cf14e1be"
		gcpRouterId := "router-eb38cf14e1be"

		kcpsubscription.Ignore.AddName(secretBindingName)
		kcpvpcnetwork.Ignore.AddName(vpcNetworkName)

		gcpMock := infra.GcpMock2().NewSubscription("rt")
		defer gcpMock.Delete()

		By("Given Subscription is created", func() {
			subscription = cloudcontrolv1beta1.NewSubscriptionBuilder().
				WithName(secretBindingName).
				WithFinalizer(api.CommonFinalizerDeletionHook).
				WithLabel(cloudcontrolv1beta1.SubscriptionLabelBindingName, secretBindingName).
				WithBindingName(secretBindingName).
				Build()

			Expect(CreateObj(infra.Ctx(), infra.KCP().Client(), subscription)).
				To(Succeed())

			err := composed.NewStatusPatcher(subscription).
				MutateStatus(func(sub *cloudcontrolv1beta1.Subscription) {
					sub.SetStatusReady()
					sub.Status.Provider = cloudcontrolv1beta1.ProviderGCP
					sub.Status.SubscriptionInfo = &cloudcontrolv1beta1.SubscriptionInfo{
						Gcp: &cloudcontrolv1beta1.SubscriptionInfoGcp{
							Project: gcpMock.ProjectId(),
						},
					}
				}).
				Patch(infra.Ctx(), infra.KCP().Client())
			Expect(err).NotTo(HaveOccurred())
		})

		By("Given VpcNetwork is created", func() {
			vpcNetwork = cloudcontrolv1beta1.NewVpcNetworkBuilder().
				WithName(vpcNetworkName).
				WithFinalizer(api.CommonFinalizerDeletionHook).
				WithType(cloudcontrolv1beta1.VpcNetworkTypeKyma).
				WithCidrBlocks("10.250.0.0/16").
				WithSubscription(subscription.Name).
				WithRegion("europe-west1").
				Build()

			Expect(CreateObj(infra.Ctx(), infra.KCP().Client(), vpcNetwork)).
				To(Succeed())

			err := composed.NewStatusPatcher(vpcNetwork).
				MutateStatus(func(n *cloudcontrolv1beta1.VpcNetwork) {
					n.SetStatusProvisioned()
					n.Status.Identifiers.Name = common.KymaVpcName(vpcNetwork.Name)
					n.Status.Identifiers.Vpc = gcpVpcNetworkId
					n.Status.Identifiers.Router = gcpRouterId
				}).
				Patch(infra.Ctx(), infra.KCP().Client())
			Expect(err).NotTo(HaveOccurred())
		})

		By("When Runtime is created", func() {
			runtime = infrastructuremanagerv1.NewSimpleRuntimeBuilder().
				WithName(name).
				WithProvider(cloudcontrolv1beta1.ProviderGCP).
				WithShootName(shootName).
				WithBindingName(secretBindingName).
				WithVpcNetworkName(new(vpcNetworkName)).
				Build()

			Expect(CreateObj(infra.Ctx(), infra.KCP().Client(), runtime)).
				To(Succeed())

			_, err := composed.PatchObjAddFinalizer(infra.Ctx(), api.CommonFinalizerDeletionHook, runtime, infra.KCP().Client())
			Expect(err).NotTo(HaveOccurred())
		})

		By("Then Runtime is successfully reconciled", func() {
			Eventually(kcpruntime.Tracker.IsReconciledWith).
				WithArguments(runtime.Name, composed.ReconciliationLabelSuccess).
				Should(Succeed())
		})

		if feature.RuntimeSecurityGcp.Value(infra.Ctx()) {

			By("Then Runtime is annotated as security handled", func() {
				Eventually(LoadAndCheck).
					WithArguments(infra.Ctx(), infra.KCP().Client(), runtime, NewObjActions(), HavingAnnotation(cloudcontrolv1beta1.RuntimeSecurityStatusAnnotation, "Ready")).
					Should(Succeed())
			})

			By("When Runtime security is enabled", func() {
				Expect(LoadAndCheck(infra.Ctx(), infra.KCP().Client(), runtime, NewObjActions())).
					To(Succeed())

				// reset the tracker so next reconciliation can be tracked
				kcpruntime.Tracker.Clear(runtime.Name)

				_, err := composed.PatchObjMergeLabel(infra.Ctx(), runtime, infra.KCP().Client(), common.TmpRuntimeSecurityEnabledLabel, "true")
				Expect(err).NotTo(HaveOccurred())
			})

			By("Then Runtime is successfully reconciled", func() {
				Eventually(kcpruntime.Tracker.IsReconciledWith).
					WithArguments(runtime.Name, composed.ReconciliationLabelSuccess).
					Should(Succeed())
			})

			By("Then GCP SCC service security-health-analytics is ENABLED", func() {
				Eventually(gcpSccServiceHasState).
					WithArguments(infra.Ctx(), gcpMock, "security-health-analytics", securitycentermanagementpb.SecurityCenterService_ENABLED).
					Should(Succeed())
			})

			By("Then GCP SCC service web-security-scanner is ENABLED", func() {
				Eventually(gcpSccServiceHasState).
					WithArguments(infra.Ctx(), gcpMock, "web-security-scanner", securitycentermanagementpb.SecurityCenterService_ENABLED).
					Should(Succeed())
			})

			By("Then GCP SCC service event-threat-detection is ENABLED", func() {
				Eventually(gcpSccServiceHasState).
					WithArguments(infra.Ctx(), gcpMock, "event-threat-detection", securitycentermanagementpb.SecurityCenterService_ENABLED).
					Should(Succeed())
			})

			By("Then GCP SCC service vm-threat-detection is ENABLED", func() {
				Eventually(gcpSccServiceHasState).
					WithArguments(infra.Ctx(), gcpMock, "vm-threat-detection", securitycentermanagementpb.SecurityCenterService_ENABLED).
					Should(Succeed())
			})

			By("Then GCP SCC service GCE_VULNERABILITY_ASSESSMENT is ENABLED", func() {
				Eventually(gcpSccServiceHasState).
					WithArguments(infra.Ctx(), gcpMock, "GCE_VULNERABILITY_ASSESSMENT", securitycentermanagementpb.SecurityCenterService_ENABLED).
					Should(Succeed())
			})

		} // if security feature enabled for GCP

		// DELETE ===============================================================

		By("When Runtime is deleted", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), runtime)).To(Succeed())
		})

		if feature.RuntimeSecurityGcp.Value(infra.Ctx()) {

			By("Then GCP SCC service security-health-analytics is DISABLED", func() {
				Eventually(gcpSccServiceHasState).
					WithArguments(infra.Ctx(), gcpMock, "security-health-analytics", securitycentermanagementpb.SecurityCenterService_DISABLED).
					Should(Succeed())
			})

			By("Then GCP SCC service web-security-scanner is DISABLED", func() {
				Eventually(gcpSccServiceHasState).
					WithArguments(infra.Ctx(), gcpMock, "web-security-scanner", securitycentermanagementpb.SecurityCenterService_DISABLED).
					Should(Succeed())
			})

			By("Then GCP SCC service event-threat-detection is DISABLED", func() {
				Eventually(gcpSccServiceHasState).
					WithArguments(infra.Ctx(), gcpMock, "event-threat-detection", securitycentermanagementpb.SecurityCenterService_DISABLED).
					Should(Succeed())
			})

			By("Then GCP SCC service vm-threat-detection is DISABLED", func() {
				Eventually(gcpSccServiceHasState).
					WithArguments(infra.Ctx(), gcpMock, "vm-threat-detection", securitycentermanagementpb.SecurityCenterService_DISABLED).
					Should(Succeed())
			})

			By("Then GCP SCC service GCE_VULNERABILITY_ASSESSMENT is DISABLED", func() {
				Eventually(gcpSccServiceHasState).
					WithArguments(infra.Ctx(), gcpMock, "GCE_VULNERABILITY_ASSESSMENT", securitycentermanagementpb.SecurityCenterService_DISABLED).
					Should(Succeed())
			})

		} // if security feature enabled for GCP

		By("Then VpcNetwork is not deleted", func() {
			Expect(LoadAndCheck(infra.Ctx(), infra.KCP().Client(), vpcNetwork, NewObjActions())).
				To(Succeed())
			Expect(vpcNetwork.DeletionTimestamp).To(BeNil(), "vpcNetwork should not be marked for deletion")
		})

		By("And Then Subscription is not deleted", func() {
			Expect(LoadAndCheck(infra.Ctx(), infra.KCP().Client(), subscription, NewObjActions())).
				To(Succeed())
			Expect(subscription.DeletionTimestamp).To(BeNil(), "subscription should not be marked for deletion")
		})

		By("// cleanup: delete Runtime", func() {
			_, err := composed.PatchObjRemoveFinalizer(infra.Ctx(), api.CommonFinalizerDeletionHook, runtime, infra.KCP().Client())
			Expect(err).NotTo(HaveOccurred())
			Expect(IsDeleted(infra.Ctx(), infra.KCP().Client(), runtime)).
				To(Succeed())
		})

		By("// cleanup: delete VpcNetwork", func() {
			_, err := composed.PatchObjRemoveFinalizer(infra.Ctx(), api.CommonFinalizerDeletionHook, vpcNetwork, infra.KCP().Client())
			Expect(err).NotTo(HaveOccurred())
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), vpcNetwork)).To(Succeed())
			Expect(IsDeleted(infra.Ctx(), infra.KCP().Client(), vpcNetwork)).
				To(Succeed())
		})

		By("// cleanup: delete Subscription", func() {
			_, err := composed.PatchObjRemoveFinalizer(infra.Ctx(), api.CommonFinalizerDeletionHook, subscription, infra.KCP().Client())
			Expect(err).NotTo(HaveOccurred())
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), subscription)).To(Succeed())
			Expect(IsDeleted(infra.Ctx(), infra.KCP().Client(), subscription)).
				To(Succeed())
		})
	})
})
