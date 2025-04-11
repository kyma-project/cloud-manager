package looper

import (
	"context"
	"github.com/go-logr/logr"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func TestInstaller(t *testing.T) {

	prepare := func(ctx context.Context) (*SkrStatus, *installer, *runtime.Scheme, Cluster) {
		skrStatus := NewSkrStatus(ctx)
		instlr := &installer{
			skrStatus:        skrStatus,
			skrProvidersPath: "../../../../config/dist/skr/crd/bases/providers",
			logger:           logr.Discard(),
		}

		scheme := runtime.NewScheme()
		assert.NoError(t, cloudresourcesv1beta1.AddToScheme(scheme))
		assert.NoError(t, clientgoscheme.AddToScheme(scheme))

		clnt := fake.NewFakeClient()
		clstr := NewCluster(scheme, clnt, clnt)

		return skrStatus, instlr, scheme, clstr
	}

	run := func(ctx context.Context, t *testing.T, provider cloudcontrolv1beta1.ProviderType, testCases []SkrStatusTestCase) {
		skrStatus, instlr, _, clstr := prepare(ctx)
		assert.NoError(t, instlr.Handle(ctx, string(provider), clstr))

		checker := NewSkrStatusChecker(skrStatus).InstallerManifest()
		checker.CheckAll(t, testCases)

		uncheker := checker.Unchecked()
		assert.Equal(t, 0, uncheker.Len(), "Unchecked handles:\n"+uncheker.String())
	}

	t.Run("aws", func(t *testing.T) {
		run(context.Background(), t, cloudcontrolv1beta1.ProviderAws, []SkrStatusTestCase{
			{"awsnfsbackupschedule.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"Creating"}},
			{"awsnfsvolumebackup.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"Creating"}},
			{"awsnfsvolumerestore.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"Creating"}},
			{"awsnfsvolume.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"Creating"}},
			{"awsrediscluster.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"Creating"}},
			{"awsredisinstance.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"Creating"}},
			{"awsvpcpeering.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"Creating"}},
			{"iprange.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"Creating"}},

			{"awsnfsbackupschedule.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			{"awsnfsvolumebackup.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			{"awsnfsvolumerestore.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			{"awsnfsvolume.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			//{"XXXawsrediscluster.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			{"awsredisinstance.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			{"awsvpcpeering.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			{"iprange.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
		})
	})

	t.Run("azure", func(t *testing.T) {
		run(context.Background(), t, cloudcontrolv1beta1.ProviderAzure, []SkrStatusTestCase{
			{"azurerwxbackupschedule.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"Creating"}},
			{"azurerwxvolumebackup.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"Creating"}},
			{"azurerwxvolumerestore.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"Creating"}},
			//{"azurerwxvolume.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"Creating"}},
			{"azurerediscluster.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"Creating"}},
			{"azureredisinstance.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"Creating"}},
			{"azurevpcpeering.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"Creating"}},
			{"iprange.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"Creating"}},

			{"azurerwxbackupschedule.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			//{"azurerwxvolumebackup.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			{"azurerwxvolumerestore.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			//{"azurerwxvolume.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			// {"azurerediscluster.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			{"azureredisinstance.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			{"azurevpcpeering.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			{"iprange.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
		})
	})

	t.Run("gcp", func(t *testing.T) {
		run(context.Background(), t, cloudcontrolv1beta1.ProviderGCP, []SkrStatusTestCase{
			{"gcpnfsbackupschedule.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"Creating"}},
			{"gcpnfsvolumebackup.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"Creating"}},
			{"gcpnfsvolumerestore.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"Creating"}},
			{"gcpnfsvolume.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"Creating"}},
			//{"gcprediscluster.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"Creating"}},
			{"gcpredisinstance.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"Creating"}},
			{"gcpvpcpeering.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"Creating"}},
			{"iprange.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"Creating"}},

			{"gcpnfsbackupschedule.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			{"gcpnfsvolumebackup.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			{"gcpnfsvolumerestore.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			{"gcpnfsvolume.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			//{"gcprediscluster.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			{"gcpredisinstance.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			{"gcpvpcpeering.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			{"iprange.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
		})
	})

	t.Run("openstack", func(t *testing.T) {
		run(context.Background(), t, cloudcontrolv1beta1.ProviderOpenStack, []SkrStatusTestCase{
			{"cceenfsvolume.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"Creating"}},
		})
	})
}
