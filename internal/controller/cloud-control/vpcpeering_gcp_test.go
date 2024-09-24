package cloudcontrol

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	networkPkg "github.com/kyma-project/cloud-manager/pkg/kcp/network"
	scopePkg "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	"github.com/kyma-project/cloud-manager/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/klog/v2"
)

var _ = Describe("Feature: KCP VpcPeering", Focus, func() {
	It("Scenario: KCP GCP VpcPeering is created", func() {
		const (
			kymaName           = "57bc9639-d752-4f67-8b9e-7cd12514575f"
			kymaNetworkName    = "57bc9639-d752-4f67-8b9e-7cd12514575f--kyma"
			remoteNetworkName  = "f5331c29-bb1a-439c-8376-94be50232eb4"
			remotePeeringName  = "peering-sap-gcp-skr-dev-cust-00002-to-sap-sc-learn"
			remoteVpc          = "default"
			remoteProject      = "sap-sc-learn"
			remoteRefNamespace = "kcp-system"
			remoteRefName      = "skr-gcp-vpcpeering"
			importCustomRoutes = false
		)

		kymaCR := util.NewKymaUnstructured()
		By("And Given Kyma CR exists", func() {
			Eventually(CreateKymaCR).
				WithArguments(infra.Ctx(), infra, kymaCR, WithName(kymaName), WithKymaSpecChannel("dev")).
				Should(Succeed(), "failed creating kyma cr")
		})

		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			scopePkg.Ignore.AddName(kymaName)

			Eventually(CreateScopeGcp).
				WithArguments(infra.Ctx(), infra, scope, WithName(kymaName)).
				Should(Succeed())
		})

		// and Given the Kyma network object exists in KCP
		kymaNetwork := cloudcontrolv1beta1.Network{
			Spec: cloudcontrolv1beta1.NetworkSpec{
				Network: cloudcontrolv1beta1.NetworkInfo{
					Managed: &cloudcontrolv1beta1.ManagedNetwork{
						Cidr:     "10.10.10.0/24",
						Location: "eu-west1",
					},
				},
				Type: cloudcontrolv1beta1.NetworkTypeKyma,
			},
		}

		By("And Given Kyma Network exists in KCP", func() {
			// Tell Scope reconciler to ignore this kymaName
			networkPkg.Ignore.AddName(kymaNetwork.Name)

			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), &kymaNetwork, WithName(kymaNetworkName), WithScope(scope.Name)).
				Should(Succeed())
		})

		// and Given the remote network object exists in KCP
		remoteNetwork := cloudcontrolv1beta1.Network{
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
			networkPkg.Ignore.AddName(remoteNetwork.Name)

			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), remoteNetwork, WithName(remoteNetwork.Name), WithScope(scope.Name)).
				Should(Succeed())
		})

		vpcpeering := &cloudcontrolv1beta1.VpcPeering{
			Spec: cloudcontrolv1beta1.VpcPeeringSpec{
				VpcPeering: &cloudcontrolv1beta1.VpcPeeringInfo{
					Gcp: &cloudcontrolv1beta1.GcpVpcPeering{
						RemotePeeringName:  remotePeeringName,
						RemoteProject:      remoteProject,
						RemoteVpc:          remoteVpc,
						ImportCustomRoutes: importCustomRoutes,
					},
				},
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
					LocalPeeringName:   "cm--" + remoteNetworkName,
					ImportCustomRoutes: false,
				},
			},
		}

		By("When KCP VpcPeering is created", func() {
			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), vpcpeering,
					WithName(remoteNetworkName),
					WithRemoteRef(remoteRefName),
					WithScope(kymaName),
				).
				Should(Succeed())
		})

		//then gcp vpc peering is created on remote side
		//and then gcp vpc peering is created on kyma side
		//when remote side vpc peering is active (call the function to set to active)
		//and when vpc peering kyma side is active (call the function to set to active)

		By("Then KCP VpcPeering has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), vpcpeering,
					NewObjActions(),
					HaveFinalizer(cloudcontrolv1beta1.FinalizerName),
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

	})
})
