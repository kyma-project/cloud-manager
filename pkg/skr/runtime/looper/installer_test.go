package looper

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	commonscheme "github.com/kyma-project/cloud-manager/pkg/common/scheme"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestInstaller(t *testing.T) {

	prepare := func(ctx context.Context) (*SkrStatus, *installer, *runtime.Scheme, Cluster) {
		skrStatus := NewSkrStatus(ctx)
		instlr := &installer{
			skrStatus:        skrStatus,
			skrProvidersPath: "../../../../config/dist/skr/crd/bases/providers",
			logger:           logr.Discard(),
		}

		clnt := fake.NewFakeClient()
		clstr := NewCluster(commonscheme.SkrScheme, clnt, clnt)

		return skrStatus, instlr, commonscheme.SkrScheme, clstr
	}

	run := func(ctx context.Context, t *testing.T, provider cloudcontrolv1beta1.ProviderType, testCases []SkrStatusTestCase) {
		skrStatus, instlr, _, clstr := prepare(ctx)
		assert.NoError(t, instlr.Handle(ctx, string(provider), clstr))

		checker := NewSkrStatusChecker(skrStatus).InstallerManifest()
		checker.CheckAll(t, testCases)

		uncheker := checker.Unchecked()
		assert.Equal(t, 0, uncheker.Len(), "Unchecked handles:\n this manifest was installed but test didnt assert them:"+uncheker.String())
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
			{"wafpolicy.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"Creating"}},
			{"certificate.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"Creating"}},

			{"awsnfsbackupschedule.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			{"awsnfsvolumebackup.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			{"awsnfsvolumerestore.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			{"awsnfsvolume.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			{"awsrediscluster.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			{"awsredisinstance.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			{"awsvpcpeering.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			{"iprange.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			{"wafpolicy.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			{"certificate.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
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
			{"azuremanagedredis.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"Creating"}},
			{"azurevpcpeering.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"Creating"}},
			{"azurevpcdnslink.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"Creating"}},
			{"iprange.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"Creating"}},

			{"azurerwxbackupschedule.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			//{"azurerwxvolumebackup.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			{"azurerwxvolumerestore.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			//{"azurerwxvolume.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			{"azurerediscluster.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			{"azureredisinstance.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			//{"azuremanagedredis.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			{"azurevpcpeering.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			{"azurevpcdnslink.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			{"iprange.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
		})
	})

	t.Run("gcp", func(t *testing.T) {
		run(context.Background(), t, cloudcontrolv1beta1.ProviderGCP, []SkrStatusTestCase{
			{"gcpnfsbackupschedule.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"Creating"}},
			{"gcpnfsvolumebackup.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"Creating"}},
			{"gcpnfsvolumebackupdiscovery.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"Creating"}},
			{"gcpnfsvolumerestore.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"Creating"}},
			{"gcpnfsvolume.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"Creating"}},
			{"gcprediscluster.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"Creating"}},
			{"gcpredisinstance.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"Creating"}},
			{"gcpvpcpeering.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"Creating"}},
			{"iprange.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"Creating"}},
			{"gcpsubnet.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"Creating"}},

			{"gcpnfsbackupschedule.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			{"gcpnfsvolumebackup.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			// {"gcpnfsvolumebackupdiscovery.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			{"gcpnfsvolumerestore.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			{"gcpnfsvolume.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			{"gcprediscluster.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			{"gcpredisinstance.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			{"gcpsubnet.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			{"gcpvpcpeering.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			{"iprange.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
		})
	})

	t.Run("openstack", func(t *testing.T) {
		run(context.Background(), t, cloudcontrolv1beta1.ProviderOpenStack, []SkrStatusTestCase{
			{"iprange.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"Creating"}},
			{"sapnfsvolume.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"Creating"}},
			{"sapnfsvolumesnapshot.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"Creating"}},
			{"sapnfsvolumesnapshotrestore.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"Creating"}},
			{"sapnfsvolumesnapshotschedule.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"Creating"}},
			{"iprange.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			{"sapnfsvolume.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			{"sapnfsvolumesnapshot.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			{"sapnfsvolumesnapshotrestore.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			{"sapnfsvolumesnapshotschedule.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
		})
	})

	t.Run("alicloud", func(t *testing.T) {
		run(context.Background(), t, cloudcontrolv1beta1.ProviderAlicloud, []SkrStatusTestCase{
			{"iprange.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"Creating"}},
			{"iprange.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			{"alicloudredisinstance.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"Creating"}},
			{"alicloudrediscluster.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"Creating"}},
			{"alicloudnfsvolume.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"Creating"}},
			// TODO: enable Busola entries once the Busola forms are designed for AliCloud Redis.
			// {"alicloudredisinstance.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			// {"alicloudrediscluster.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
		})
	})

	// aws_apiNotEnabled uses the non-parity fixture (apiEnabled=false).
	// The guard feature.ApiEnabled.Value() = false → ApiNotEnabled.
	// Recording "ApiNotEnabled" proves the installer reads apiEnabled.
	//
	// Every feature-tagged AWS CRD records ApiNotEnabled; iprange (FeatureFoundation)
	// stays Creating via the always-on carve-out (isAlwaysEnabledApi). Asserting the full
	// manifest with the Unchecked() guard (via run) proves no CRD leaks through the gate
	// and that the foundation carve-out holds while every feature API goes dark.
	t.Run("aws_apiNotEnabled", func(t *testing.T) {
		require.NoError(t, feature.Initialize(context.Background(), logr.Discard(),
			feature.WithFile("testdata/apiEnabledNonParity.yaml")))
		t.Cleanup(func() {
			if err := feature.Initialize(context.Background(), logr.Discard(),
				feature.WithFile("testdata/ffAllEnabled.yaml")); err != nil {
				t.Errorf("cleanup: restore ffAllEnabled.yaml: %v", err)
			}
		})

		run(context.Background(), t, cloudcontrolv1beta1.ProviderAws, []SkrStatusTestCase{
			{"awsnfsbackupschedule.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"ApiNotEnabled"}},
			{"awsnfsvolumebackup.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"ApiNotEnabled"}},
			{"awsnfsvolumerestore.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"ApiNotEnabled"}},
			{"awsnfsvolume.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"ApiNotEnabled"}},
			{"awsrediscluster.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"ApiNotEnabled"}},
			{"awsredisinstance.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"ApiNotEnabled"}},
			{"awsvpcpeering.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"ApiNotEnabled"}},
			{"iprange.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"Creating"}},
			{"wafpolicy.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"ApiNotEnabled"}},
			{"certificate.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormCrd, []string{"ApiNotEnabled"}},

			{"awsnfsbackupschedule.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"ApiNotEnabled"}},
			{"awsnfsvolumebackup.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"ApiNotEnabled"}},
			{"awsnfsvolumerestore.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"ApiNotEnabled"}},
			{"awsnfsvolume.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"ApiNotEnabled"}},
			{"awsrediscluster.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"ApiNotEnabled"}},
			{"awsredisinstance.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"ApiNotEnabled"}},
			{"awsvpcpeering.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"ApiNotEnabled"}},
			{"iprange.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"Creating"}},
			{"wafpolicy.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"ApiNotEnabled"}},
			{"certificate.cloud-resources.kyma-project.io", true, "InstallerManifest", KindFormBusola, []string{"ApiNotEnabled"}},
		})
	})
}
