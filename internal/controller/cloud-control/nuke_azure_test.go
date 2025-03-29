package cloudcontrol

import (
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	scopePkg "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const fileSharePattern = "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.RecoveryServices/vaults/%s/backupFabrics/Azure/protectionContainers/StorageContainer;Storage;%s;%s/protectedItems/AzureFileShare;%s"

var _ = Describe("Feature: KCP Nuke AzureRwxVolumeBackup", func() {

	//Define variables
	scopeName := "test-nuke-azure-rwx-backups-scope-01"
	scope := &cloudcontrolv1beta1.Scope{}

	vaultName := fmt.Sprintf("cm-%s", scopeName)
	location := "westeurope"
	saName := "test-sa"
	rgName := "test-rg"
	backupPolicyName := "test-policy"
	containerName := fmt.Sprintf("StorageContainer;Storage;%s;%s", rgName, saName)
	fileShares := []string{"test-fs-01", "test-fs-02"}

	BeforeEach(func() {
		By("Given KCP Scope exists", func() {
			scopePkg.Ignore.AddName(scopeName)
			// Given Scope exists
			Eventually(GivenScopeAzureExists).
				WithArguments(
					infra.Ctx(), infra, scope,
					WithName(scopeName),
				).
				Should(Succeed())
		})
		By("And Given Scope is in Ready state", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), scope,
					NewObjActions(),
				).
				Should(Succeed())

			//Update KCP Scope status to Ready
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), scope,
					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed())
		})

		clientProvider := infra.AzureMock().StorageProvider()
		subscriptionId, tenantId := scope.Spec.Scope.Azure.SubscriptionId, scope.Spec.Scope.Azure.TenantId
		nukeClient, _ := clientProvider(infra.Ctx(), "", "", subscriptionId, tenantId, "")
		By(" And Given azure Vault exits for the same scope", func() {

			_, err := nukeClient.CreateVault(
				infra.Ctx(), rgName, vaultName, location)
			Expect(err).
				ShouldNot(HaveOccurred(), "failed creating azure Vault directly")
		})

		By(" And Given azure Backups exits for the same scope", func() {

			for _, fileShareName := range fileShares {
				err := nukeClient.CreateOrUpdateProtectedItem(
					infra.Ctx(), subscriptionId, location, vaultName, rgName, containerName, fileShareName, backupPolicyName, saName)
				Expect(err).ShouldNot(HaveOccurred(), "failed creating Azure Recovery Point directly")
			}
		})
	})

	//Define variables.
	nukeName := "nuke-" + scopeName
	nuke := &cloudcontrolv1beta1.Nuke{}
	It("When Nuke for the Scope is created", func() {
		//TODO: Remove skip when the test passes.
		Skip("Nuke Backups for Azure is disabled")

		clientProvider := infra.AzureMock().StorageProvider()
		subscriptionId, tenantId := scope.Spec.Scope.Azure.SubscriptionId, scope.Spec.Scope.Azure.TenantId
		nukeClient, _ := clientProvider(infra.Ctx(), "", "", subscriptionId, tenantId, "")

		Eventually(CreateObj(infra.Ctx(), infra.KCP().Client(), nuke,
			WithName(nukeName),
			WithScope(scopeName),
		)).Should(Succeed())

		By("Then Nuke status state is Deleting", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), nuke, NewObjActions(),
					HavingState("Completed"),
				).
				Should(Succeed())
		})

		kind := "AzureRwxVolumeBackup"

		By(fmt.Sprintf("And Then provider resource %s does not exist", kind), func() {
			Eventually(func() error {
				backupsOnAzure, err := nukeClient.ListProtectedItems(infra.Ctx(), vaultName, rgName)
				if err != nil {
					return err
				}
				Expect(backupsOnAzure).To(HaveLen(0))
				return nil
			}).Should(Succeed())
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
				for _, fileShareName := range fileShares {
					id := fmt.Sprintf(fileSharePattern, subscriptionId, rgName, vaultName, rgName, saName, fileShareName)
					actual := sk.Objects[id]
					if actual == cloudcontrolv1beta1.NukeResourceStatusDeleted {
						continue
					}
					return fmt.Errorf("expected resource %s/%s to have status Deleted, but found %s", kind, id, actual)
				}
				return nil
			}).Should(Succeed())
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
