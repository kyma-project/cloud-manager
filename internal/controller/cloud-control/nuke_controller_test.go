package cloudcontrol

import (
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	kcpiprange "github.com/kyma-project/cloud-manager/pkg/kcp/iprange"
	kcpnetwork "github.com/kyma-project/cloud-manager/pkg/kcp/network"
	kcpnfsinstance "github.com/kyma-project/cloud-manager/pkg/kcp/nfsinstance"
	kcpredisinstance "github.com/kyma-project/cloud-manager/pkg/kcp/redisinstance"
	scopePkg "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	kcpvpcpeering "github.com/kyma-project/cloud-manager/pkg/kcp/vpcpeering"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Feature: Cleanup orphan resources", func() {

	It("Scenario: KCP Nuke deletes all resources in the Scope", func() {
		const kymaName = "5ee3f7f2-fd7e-4718-abd2-ff9f5783f0fd"

		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			scopePkg.Ignore.AddName(kymaName)

			Expect(CreateScopeAzure(infra.Ctx(), infra, scope, WithName(kymaName))).
				To(Succeed(), "failed creating scope")
		})

		cmNetwork := cloudcontrolv1beta1.NewNetworkBuilder().
			WithName("b547afe6-80a4-45c3-8be7-6b95f828ca1a").
			WithType(cloudcontrolv1beta1.NetworkTypeCloudResources).
			WithScope(kymaName).
			WithCidr(common.DefaultCloudManagerCidr).
			Build()

		By("And Given CM Network exists", func() {
			kcpnetwork.Ignore.AddName(cmNetwork.Name)

			Expect(CreateObj(infra.Ctx(), infra.KCP().Client(), cmNetwork,
				AddFinalizer(cloudcontrolv1beta1.FinalizerName),
			)).To(Succeed(), "failed creating cm network")
		})

		ipRangeName := "86cdafcd-5816-48a0-8d5d-c936f2a15f06"
		ipRange := &cloudcontrolv1beta1.IpRange{}

		By("And Given IpRange exists", func() {
			kcpiprange.Ignore.AddName(ipRangeName)

			Eventually(CreateKcpIpRange).
				WithArguments(infra.Ctx(), infra.KCP().Client(), ipRange,
					WithName(ipRangeName),
					AddFinalizer(cloudcontrolv1beta1.FinalizerName),
					WithKcpIpRangeNetwork(cmNetwork.Name),
					WithScope(kymaName),
					WithRemoteRef("foo"),
					WithKcpIpRangeSpecCidr(common.DefaultCloudManagerCidr),
				).
				Should(Succeed(), "failed creating IpRange")
		})

		vpcPeering := cloudcontrolv1beta1.NewVpcPeeringBuilder().
			WithName("76b8fbba-588f-4662-b8d0-f5e9beacad47").
			WithScope(kymaName).
			WithRemoteRef(DefaultSkrNamespace, "name").
			WithAzurePeering("remotePeeringName", "remoteVNet", "remoteResourceGroup").
			Build()

		By("And Given VpcPeering exists", func() {
			kcpvpcpeering.Ignore.AddName(vpcPeering.Name)

			Expect(CreateObj(infra.Ctx(), infra.KCP().Client(), vpcPeering,
				AddFinalizer(cloudcontrolv1beta1.FinalizerName),
			)).To(Succeed(), "failed creating VpcPeering")
		})

		redisInstanceName := "bcdafb60-efa6-4996-8bd4-3726959fd1e1"
		redisInstance := &cloudcontrolv1beta1.RedisInstance{}

		By("And Given RedisInstance exists", func() {
			kcpredisinstance.Ignore.AddName(redisInstanceName)

			Expect(CreateRedisInstance(infra.Ctx(), infra.KCP().Client(), redisInstance,
				WithName(redisInstanceName),
				AddFinalizer(cloudcontrolv1beta1.FinalizerName),
				WithRemoteRef("remote-redis"),
				WithIpRange(ipRange.Name),
				WithScope(kymaName),
				WithRedisInstanceAzure(),
				WithSKU(2),
				WithKcpAzureRedisVersion("6.0"),
			)).To(Succeed(), "failed creating RedisInstance")
		})

		nfsInstanceName := "59b4c25d-b90f-4641-9837-9623b910f264"
		nfsInstance := &cloudcontrolv1beta1.NfsInstance{}

		By("And Given NfsInstance exists", func() {
			kcpnfsinstance.Ignore.AddName(nfsInstanceName)

			Expect(CreateNfsInstance(infra.Ctx(), infra.KCP().Client(), nfsInstance,
				WithName(nfsInstanceName),
				AddFinalizer(cloudcontrolv1beta1.FinalizerName),
				WithRemoteRef("foo"),
				WithScope(kymaName),
				WithIpRange(ipRange.Name),
				WithNfsInstanceAws(), // never mind it doesn't match Azure, won't be reconciled anyway
				AddFinalizer(cloudcontrolv1beta1.FinalizerName),
			)).To(Succeed(), "failed creating NfsInstance")
		})

		nuke := &cloudcontrolv1beta1.Nuke{}

		By("When Nuke for the Scope is created", func() {
			Expect(CreateObj(infra.Ctx(), infra.KCP().Client(), nuke,
				WithName("e03551a7-bf7e-4d6a-b9ba-3415d3f3401f"),
				WithScope(kymaName),
			)).To(Succeed())
		})

		By("Then Nuke status state is Deleting", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), nuke, NewObjActions(),
					HavingState("Deleting"),
				).
				Should(Succeed())
		})

		resources := map[string]focal.CommonObject{
			"VpcPeering":    vpcPeering,
			"RedisInstance": redisInstance,
			"NfsInstance":   nfsInstance,
			"IpRange":       ipRange,
			"Network":       cmNetwork,
		}

		for kind, obj := range resources {
			By(fmt.Sprintf("And Then Nuke status resource %s has Deleting status", kind), func() {
				sk := nuke.Status.GetKindNoCreate(kind)
				Expect(sk).NotTo(BeNil())
				Expect(sk.Kind).To(Equal(kind))
				Expect(sk.Objects).To(HaveLen(1))
				Expect(sk.Objects).To(HaveKeyWithValue(obj.GetName(), cloudcontrolv1beta1.NukeResourceStatusDeleting))
			})

			By(fmt.Sprintf("And Then resource %s has deletion timestamp", kind), func() {
				Expect(LoadAndCheck(infra.Ctx(), infra.KCP().Client(), obj, NewObjActions())).
					To(Succeed())
				Expect(obj.GetDeletionTimestamp()).NotTo(BeNil())
			})
		}

		for kind, obj := range resources {
			By(fmt.Sprintf("When resource %s finalizer is removed", kind), func() {
				removed, err := composed.PatchObjRemoveFinalizer(infra.Ctx(), cloudcontrolv1beta1.FinalizerName, obj, infra.KCP().Client())
				Expect(err).To(Succeed())
				Expect(removed).To(BeTrue())
			})

			By(fmt.Sprintf("Then resource %s does not exist", kind), func() {
				Eventually(IsDeleted).
					WithArguments(infra.Ctx(), infra.KCP().Client(), obj).
					Should(Succeed())
			})

			By(fmt.Sprintf("And Then Nuke status resource %s has state Deleted", kind), func() {
				Eventually(func() error {
					if err := LoadAndCheck(infra.Ctx(), infra.KCP().Client(), nuke, NewObjActions()); err != nil {
						return err
					}
					sk := nuke.Status.GetKindNoCreate(kind)
					if sk == nil {
						return fmt.Errorf("kind %s not found in Nuke status", kind)
					}
					actual := sk.Objects[obj.GetName()]
					if actual == cloudcontrolv1beta1.NukeResourceStatusDeleted {
						return nil
					}
					return fmt.Errorf("expected resource %s/%s to have status Deleted, but found %s", kind, obj.GetName(), actual)
				}).Should(Succeed())
			})
		}

		By("And Then Nuke status state is Completed", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), nuke, NewObjActions(),
					HavingState("Completed"),
				).Should(Succeed())
		})

		By("And Then Scope is deleted", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), scope).
				Should(Succeed())
		})

		By("// cleanup: Delete Nuke", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), nuke)).
				To(Succeed())
		})

	})
})
