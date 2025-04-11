package cloudcontrol

import (
	"errors"

	pb "cloud.google.com/go/compute/apiv1/computepb"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	kcpnetwork "github.com/kyma-project/cloud-manager/pkg/kcp/network"
	kcpscope "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
)

var _ = Describe("Feature: KCP VpcPeering", func() {
	It("Scenario: KCP GCP VpcPeering is created", func() {
		const (
			kymaName           = "7e829442-f92e-4205-9d36-0d622a422d74"
			kymaNetworkName    = kymaName + "--kyma"
			kymaProject        = "kyma-project"
			kymaVpc            = "shoot-12345-abc"
			remoteNetworkName  = "f5331c29-bb1a-439c-8376-94be50232eb4"
			remotePeeringName  = "peering-sap-gcp-skr-dev-cust-00002-to-sap-sc-learn"
			remoteVpc          = "default"
			remoteProject      = "sap-sc-learn"
			remoteRefNamespace = "kcp-system"
			remoteRefName      = "skr-gcp-vpcpeering"
			importCustomRoutes = false
		)

		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			kcpscope.Ignore.AddName(kymaName)

			Eventually(CreateScopeGcp).
				WithArguments(infra.Ctx(), infra, scope, WithName(kymaName)).
				Should(Succeed())
		})

		// and Given the Kyma network object exists in KCP
		kymaNetwork := &cloudcontrolv1beta1.Network{
			Spec: cloudcontrolv1beta1.NetworkSpec{
				Network: cloudcontrolv1beta1.NetworkInfo{
					Reference: &cloudcontrolv1beta1.NetworkReference{
						Gcp: &cloudcontrolv1beta1.GcpNetworkReference{
							GcpProject:  kymaProject,
							NetworkName: kymaVpc,
						},
					},
				},
				Type: cloudcontrolv1beta1.NetworkTypeKyma,
			},
		}

		By("And Given Kyma Network exists in KCP", func() {
			// Tell Scope reconciler to ignore this kymaName
			kcpnetwork.Ignore.AddName(kymaNetworkName)

			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kymaNetwork, WithName(kymaNetworkName), WithScope(scope.Name)).
				Should(Succeed())
		})

		// and Given the remote network object exists in KCP
		remoteNetwork := &cloudcontrolv1beta1.Network{
			Spec: cloudcontrolv1beta1.NetworkSpec{
				Network: cloudcontrolv1beta1.NetworkInfo{
					Reference: &cloudcontrolv1beta1.NetworkReference{
						Gcp: &cloudcontrolv1beta1.GcpNetworkReference{
							GcpProject:  remoteProject,
							NetworkName: remoteVpc,
						},
					},
				},
				Type: cloudcontrolv1beta1.NetworkTypeExternal,
			},
		}

		By("And Given Remote Network exists in KCP", func() {
			// Tell Scope reconciler to ignore this kymaName
			kcpnetwork.Ignore.AddName(remoteNetworkName)

			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), remoteNetwork, WithName(remoteNetworkName), WithScope(scope.Name), WithState("Ready")).
				Should(Succeed())
		})

		By("When KCP KymaNetwork is Ready", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(),
					infra.KCP().Client(),
					kymaNetwork,
					WithNetworkStatusNetwork(kymaNetwork.Spec.Network.Reference),
					WithState("Ready"),
					WithConditions(KcpReadyCondition())).
				Should(Succeed())
		})

		By("And When KCP RemoteNetwork is Ready", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(),
					infra.KCP().Client(),
					remoteNetwork,
					WithNetworkStatusNetwork(remoteNetwork.Spec.Network.Reference),
					WithState("Ready"),
					WithConditions(KcpReadyCondition())).
				Should(Succeed())
		})

		vpcpeering := &cloudcontrolv1beta1.VpcPeering{
			Spec: cloudcontrolv1beta1.VpcPeeringSpec{
				Details: &cloudcontrolv1beta1.VpcPeeringDetails{
					LocalNetwork: klog.ObjectRef{
						Name:      kymaNetwork.Name,
						Namespace: kymaNetwork.Namespace,
					},
					RemoteNetwork: klog.ObjectRef{
						Name:      remoteNetwork.Name,
						Namespace: remoteNetwork.Namespace,
					},
					PeeringName:        remotePeeringName,
					LocalPeeringName:   "cm-" + remoteNetworkName,
					ImportCustomRoutes: importCustomRoutes,
				},
			},
		}

		By("And When the remote network is tagged", func() {
			infra.GcpMock().SetMockVpcPeeringTags(remoteProject, remoteVpc, []string{kymaVpc})
		})

		By("And When the KCP VpcPeering is created", func() {
			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), vpcpeering,
					WithName(remoteNetworkName),
					WithRemoteRef(remoteRefName),
					WithScope(kymaName),
				).
				Should(Succeed())
		})

		var remotePeeringObject *pb.NetworkPeering
		By("Then GCP VpcPeering is created on remote side", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), vpcpeering,
					NewObjActions(),
					HavingVpcPeeringStatusRemoteId(),
				).
				Should(Succeed())
			remotePeeringObject = infra.GcpMock().GetMockVpcPeering(remoteProject, remoteVpc)
		})

		var kymaPeeringObject *pb.NetworkPeering
		By("And Then GCP VpcPeering is created on kyma side", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), vpcpeering,
					NewObjActions(),
					HavingVpcPeeringStatusId(),
				).
				Should(Succeed())
			kymaPeeringObject = infra.GcpMock().GetMockVpcPeering(kymaProject, kymaVpc)
		})

		By("When GCP VpcPeering is active on the remote side", func() {
			// GCP will set both to ACTIVE when the peering is active
			infra.GcpMock().SetMockVpcPeeringLifeCycleState(remoteProject, remoteVpc, pb.NetworkPeering_ACTIVE)
			Expect(ptr.Deref(remotePeeringObject.State, "") == pb.NetworkPeering_ACTIVE.String()).Should(BeTrue())
		})

		By("And When GCP VpcPeering is active on the kyma side", func() {
			// GCP will set both to ACTIVE when the peering is active
			infra.GcpMock().SetMockVpcPeeringLifeCycleState(kymaProject, kymaVpc, pb.NetworkPeering_ACTIVE)
			Expect(ptr.Deref(kymaPeeringObject.State, "") == pb.NetworkPeering_ACTIVE.String()).Should(BeTrue())
		})

		By("Then KCP VpcPeering has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), vpcpeering,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

		// DELETE
		By("When KCP VpcPeering is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), vpcpeering).
				Should(Succeed(), "Error deleting VPC Peering")
		})

		By("Then VpcPeering does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), vpcpeering).
				Should(Succeed(), "VPC Peering was not deleted")
		})

		// Check if the peering is deleted on the kyma side
	})

	It("Scenario: KCP GCP VpcPeering can be deleted due to issues on the remote network", func() {
		const (
			kymaName           = "ec697362-8f63-4423-b34f-8a99c0460d46"
			kymaNetworkName    = kymaName + "--kyma"
			kymaProject        = "kyma-project"
			kymaVpc            = "shoot-12345-abc"
			remoteNetworkName  = "0ab0eca3-3094-4842-9834-7492aaa0639d"
			remotePeeringName  = "peering-with-permission-error"
			remoteVpc          = "remote-vpc"
			remoteProject      = "remote-project"
			remoteRefNamespace = "kcp-system"
			remoteRefName      = "skr-gcp-vpcpeering"
			importCustomRoutes = false
		)

		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			kcpscope.Ignore.AddName(kymaName)

			Eventually(CreateScopeGcp).
				WithArguments(infra.Ctx(), infra, scope, WithName(kymaName)).
				Should(Succeed())
		})

		// and Given the Kyma network object exists in KCP
		kymaNetwork := &cloudcontrolv1beta1.Network{
			Spec: cloudcontrolv1beta1.NetworkSpec{
				Network: cloudcontrolv1beta1.NetworkInfo{
					Reference: &cloudcontrolv1beta1.NetworkReference{
						Gcp: &cloudcontrolv1beta1.GcpNetworkReference{
							GcpProject:  kymaProject,
							NetworkName: kymaVpc,
						},
					},
				},
				Type: cloudcontrolv1beta1.NetworkTypeKyma,
			},
		}

		By("And Given Kyma Network exists in KCP", func() {
			// Tell Scope reconciler to ignore this kymaName
			kcpnetwork.Ignore.AddName(kymaNetworkName)

			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kymaNetwork, WithName(kymaNetworkName), WithScope(scope.Name)).
				Should(Succeed())
		})

		// and Given the remote network object exists in KCP
		remoteNetwork := &cloudcontrolv1beta1.Network{
			Spec: cloudcontrolv1beta1.NetworkSpec{
				Network: cloudcontrolv1beta1.NetworkInfo{
					Reference: &cloudcontrolv1beta1.NetworkReference{
						Gcp: &cloudcontrolv1beta1.GcpNetworkReference{
							GcpProject:  remoteProject,
							NetworkName: remoteVpc,
						},
					},
				},
				Type: cloudcontrolv1beta1.NetworkTypeExternal,
			},
		}

		By("And Given Remote Network exists in KCP", func() {
			// Tell Scope reconciler to ignore this kymaName
			kcpnetwork.Ignore.AddName(remoteNetworkName)

			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), remoteNetwork, WithName(remoteNetworkName), WithScope(scope.Name), WithState("Ready")).
				Should(Succeed())
		})

		By("And Given KCP KymaNetwork is Ready", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(),
					infra.KCP().Client(),
					kymaNetwork,
					WithNetworkStatusNetwork(kymaNetwork.Spec.Network.Reference),
					WithState("Ready"),
					WithConditions(KcpReadyCondition())).
				Should(Succeed())
		})

		By("And Given KCP RemoteNetwork is Ready", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(),
					infra.KCP().Client(),
					remoteNetwork,
					WithNetworkStatusNetwork(remoteNetwork.Spec.Network.Reference),
					WithState("Ready"),
					WithConditions(KcpReadyCondition())).
				Should(Succeed())
		})

		vpcpeering := &cloudcontrolv1beta1.VpcPeering{
			Spec: cloudcontrolv1beta1.VpcPeeringSpec{
				Details: &cloudcontrolv1beta1.VpcPeeringDetails{
					LocalNetwork: klog.ObjectRef{
						Name:      kymaNetwork.Name,
						Namespace: kymaNetwork.Namespace,
					},
					RemoteNetwork: klog.ObjectRef{
						Name:      remoteNetwork.Name,
						Namespace: remoteNetwork.Namespace,
					},
					PeeringName:        remotePeeringName,
					LocalPeeringName:   "cm-" + remoteNetworkName,
					ImportCustomRoutes: importCustomRoutes,
				},
			},
		}

		By("And Given there is no permission to read remote network tags", func() {
			infra.GcpMock().SetMockVpcPeeringError(remoteProject, remoteVpc, errors.New("permission denied"))
		})

		By("And Given KCP VpcPeering is created", func() {
			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), vpcpeering,
					WithName(remoteNetworkName),
					WithRemoteRef(remoteRefName),
					WithScope(kymaName),
				).
				Should(Succeed())
		})

		By("Then KCP VpcPeering has Error condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), vpcpeering,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeError),
				).
				Should(Succeed())
		})

		By("When KCP VpcPeering in error state is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), vpcpeering).
				Should(Succeed(), "Error deleting VPC Peering")
		})

		By("Then VpcPeering does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), vpcpeering).
				Should(Succeed(), "VPC Peering was not deleted")
		})

	})

	It("Scenario: KCP GCP VpcPeering can be deleted when Networks are not in Ready state", func() {
		const (
			kymaName           = "21445c56-35fa-423a-a8d3-7bd9f3ed4976"
			kymaNetworkName    = kymaName + "--kyma"
			kymaProject        = "kyma-project"
			kymaVpc            = "shoot-12345-abc-357"
			remoteNetworkName  = "2d10d06f-81f5-4155-adae-1922a9d2dd08"
			remotePeeringName  = "peering-with-permission-deleting-with-warning"
			remoteVpc          = "remote-vpc-warning-test"
			remoteProject      = "remote-project"
			remoteRefNamespace = "kcp-system"
			remoteRefName      = "skr-gcp-vpcpeering-45"
			importCustomRoutes = false
		)

		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			kcpscope.Ignore.AddName(kymaName)

			Eventually(CreateScopeGcp).
				WithArguments(infra.Ctx(), infra, scope, WithName(kymaName)).
				Should(Succeed())
		})

		// and Given the Kyma network object exists in KCP
		kymaNetwork := &cloudcontrolv1beta1.Network{
			Spec: cloudcontrolv1beta1.NetworkSpec{
				Network: cloudcontrolv1beta1.NetworkInfo{
					Reference: &cloudcontrolv1beta1.NetworkReference{
						Gcp: &cloudcontrolv1beta1.GcpNetworkReference{
							GcpProject:  kymaProject,
							NetworkName: kymaVpc,
						},
					},
				},
				Type: cloudcontrolv1beta1.NetworkTypeKyma,
			},
		}

		By("And Given Kyma Network exists in KCP", func() {
			// Tell Scope reconciler to ignore this kymaName
			kcpnetwork.Ignore.AddName(kymaNetworkName)

			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kymaNetwork, WithName(kymaNetworkName), WithScope(scope.Name)).
				Should(Succeed())
		})

		// and Given the remote network object exists in KCP
		remoteNetwork := &cloudcontrolv1beta1.Network{
			Spec: cloudcontrolv1beta1.NetworkSpec{
				Network: cloudcontrolv1beta1.NetworkInfo{
					Reference: &cloudcontrolv1beta1.NetworkReference{
						Gcp: &cloudcontrolv1beta1.GcpNetworkReference{
							GcpProject:  remoteProject,
							NetworkName: remoteVpc,
						},
					},
				},
				Type: cloudcontrolv1beta1.NetworkTypeExternal,
			},
		}

		By("And Given Remote Network exists in KCP", func() {
			// Tell Scope reconciler to ignore this kymaName
			kcpnetwork.Ignore.AddName(remoteNetworkName)
			infra.GcpMock().SetMockVpcPeeringTags(remoteProject, remoteVpc, []string{kymaVpc})
			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), remoteNetwork, WithName(remoteNetworkName), WithScope(scope.Name), WithState("Ready")).
				Should(Succeed())
		})

		By("And Given KCP KymaNetwork is Ready", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(),
					infra.KCP().Client(),
					kymaNetwork,
					WithNetworkStatusNetwork(kymaNetwork.Spec.Network.Reference),
					WithState("Ready"),
					WithConditions(KcpReadyCondition())).
				Should(Succeed())
		})

		By("And Given KCP RemoteNetwork is Ready", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(),
					infra.KCP().Client(),
					remoteNetwork,
					WithNetworkStatusNetwork(remoteNetwork.Spec.Network.Reference),
					WithState("Ready"),
					WithConditions(KcpReadyCondition())).
				Should(Succeed())
		})

		vpcpeering := &cloudcontrolv1beta1.VpcPeering{
			Spec: cloudcontrolv1beta1.VpcPeeringSpec{
				Details: &cloudcontrolv1beta1.VpcPeeringDetails{
					LocalNetwork: klog.ObjectRef{
						Name:      kymaNetwork.Name,
						Namespace: kymaNetwork.Namespace,
					},
					RemoteNetwork: klog.ObjectRef{
						Name:      remoteNetwork.Name,
						Namespace: remoteNetwork.Namespace,
					},
					PeeringName:        remotePeeringName,
					LocalPeeringName:   "cm-" + remoteNetworkName,
					ImportCustomRoutes: importCustomRoutes,
				},
			},
		}

		By("And Given KCP VpcPeering is created", func() {
			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), vpcpeering,
					WithName(remoteNetworkName),
					WithRemoteRef(remoteRefName),
					WithScope(kymaName),
				).
				Should(Succeed())
		})

		var remotePeeringObject *pb.NetworkPeering
		By("And Given GCP VpcPeering is created on remote side", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), vpcpeering,
					NewObjActions(),
					HavingVpcPeeringStatusRemoteId(),
				).
				Should(Succeed())
			remotePeeringObject = infra.GcpMock().GetMockVpcPeering(remoteProject, remoteVpc)
		})

		var kymaPeeringObject *pb.NetworkPeering
		By("And Given GCP VpcPeering is created on kyma side", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), vpcpeering,
					NewObjActions(),
					HavingVpcPeeringStatusId(),
				).
				Should(Succeed())
			kymaPeeringObject = infra.GcpMock().GetMockVpcPeering(kymaProject, kymaVpc)
		})

		By("And given GCP VpcPeering is active on the remote side", func() {
			// GCP will set both to ACTIVE when the peering is active
			infra.GcpMock().SetMockVpcPeeringLifeCycleState(remoteProject, remoteVpc, pb.NetworkPeering_ACTIVE)
			Expect(ptr.Deref(remotePeeringObject.State, "") == pb.NetworkPeering_ACTIVE.String()).Should(BeTrue())
		})

		By("And Given GCP VpcPeering is active on the kyma side", func() {
			// GCP will set both to ACTIVE when the peering is active
			infra.GcpMock().SetMockVpcPeeringLifeCycleState(kymaProject, kymaVpc, pb.NetworkPeering_ACTIVE)
			Expect(ptr.Deref(kymaPeeringObject.State, "") == pb.NetworkPeering_ACTIVE.String()).Should(BeTrue())
		})

		By("And Given KCP VpcPeering has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), vpcpeering,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

		By("And Given KCP KymaNetwork went has Warning state", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(),
					infra.KCP().Client(),
					kymaNetwork,
					WithoutConditions("Ready"),
					WithState("Warning"),
					WithConditions(KcpWarningCondition())).
				Should(Succeed())
		})

		By("And Given KCP RemoteNetwork has Warning state", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(),
					infra.KCP().Client(),
					remoteNetwork,
					WithoutConditions("Ready"),
					WithState("Warning"),
					WithConditions(KcpWarningCondition())).
				Should(Succeed())
		})

		By("When KCP VpcPeering in deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), vpcpeering).
				Should(Succeed(), "Error deleting VPC Peering")
		})

		By("Then VpcPeering does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), vpcpeering).
				Should(Succeed(), "VPC Peering was not deleted")
		})

	})
})
