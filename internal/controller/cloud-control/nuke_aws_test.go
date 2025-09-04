package cloudcontrol

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	awsnukeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/nuke/client"
	kcpscope "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	awsnfsvolumebackupclient "github.com/kyma-project/cloud-manager/pkg/skr/awsnfsvolumebackup/client"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
)

var _ = Describe("Feature: KCP Nuke AwsNfsVolumeBackup", func() {

	//Define variables
	scopeName := "test-nuke-aws-nfs-backups-scope-01"
	scope := &cloudcontrolv1beta1.Scope{}

	vaultName := fmt.Sprintf("cm-%s", scopeName)
	clientProvider := awsnukeclient.Mock()

	recoveryPointArns := []string{"", ""}

	BeforeEach(func() {
		By("Given KCP Scope exists", func() {
			kcpscope.Ignore.AddName(scopeName)
			// Given Scope exists
			Eventually(GivenScopeAwsExists).
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

		nukeClient, _ := clientProvider(infra.Ctx(),
			scope.Spec.Scope.Aws.AccountId, scope.Spec.Region, "", "", "")

		By(" And Given Aws Vault exits for the same scope", func() {
			_, err := nukeClient.CreateBackupVault(
				infra.Ctx(),
				vaultName,
				make(map[string]string),
			)
			Expect(err).
				ShouldNot(HaveOccurred(), "failed creating Aws Vault directly")
		})

		By(" And Given Aws Backups exits for the same scope", func() {
			for i := range recoveryPointArns {
				out, err := nukeClient.StartBackupJob(
					infra.Ctx(),
					&awsnfsvolumebackupclient.StartBackupJobInput{
						BackupVaultName: vaultName,
					})
				Expect(err).ShouldNot(HaveOccurred(), "failed creating Aws Recovery Point directly")
				recoveryPointArns[i] = ptr.Deref(out.RecoveryPointArn, "")
			}
		})
	})

	//Define variables.
	nukeName := "nuke-" + scopeName
	nuke := &cloudcontrolv1beta1.Nuke{}
	It("When Nuke for the Scope is created", func() {
		//Disable the test case if the feature is not enabled.
		if !feature.FFNukeBackupsAws.Value(context.Background()) {
			Skip("Nuke Backups for AWS is disabled")
		}

		nukeClient, _ := clientProvider(infra.Ctx(),
			scope.Spec.Scope.Aws.AccountId, scope.Spec.Region, "", "", "")

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

		kind := "AwsNfsVolumeBackup"

		By(fmt.Sprintf("And Then provider resource %s does not exist", kind), func() {
			Eventually(func() error {
				backupsOnAws, err := nukeClient.ListRecoveryPointsForVault(infra.Ctx(), "", vaultName)
				if err != nil {
					return err
				}
				Expect(backupsOnAws).To(HaveLen(0))
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
				for _, arn := range recoveryPointArns {
					actual := sk.Objects[arn]
					if actual == cloudcontrolv1beta1.NukeResourceStatusDeleted {
						continue
					}
					return fmt.Errorf("expected resource %s/%s to have status Deleted, but found %s", kind, arn, actual)
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
