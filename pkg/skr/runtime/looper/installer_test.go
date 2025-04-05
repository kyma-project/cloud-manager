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

	run := func(ctx context.Context, t *testing.T, provider cloudcontrolv1beta1.ProviderType, expectedCrds []string, expectedBusola []string) {
		skrStatus, instlr, _, clstr := prepare(ctx)
		assert.NoError(t, instlr.Handle(ctx, string(provider), clstr))

		checker := NewSkrStatusChecker(skrStatus).InstallerManifest()

		for _, kind := range expectedCrds {
			checker.Crd(kind).Check(t)
		}
		for _, kind := range expectedBusola {
			checker.Busola(kind).Check(t)
		}

		uncheker := checker.Unchecked()
		assert.Equal(t, 0, uncheker.Len(), "Unchecked handles:\n"+uncheker.String())
	}

	t.Run("aws", func(t *testing.T) {
		expectedCrds := []string{
			"awsnfsbackupschedule.cloud-resources.kyma-project.io",
			"awsnfsvolumebackup.cloud-resources.kyma-project.io",
			"awsnfsvolumerestore.cloud-resources.kyma-project.io",
			"awsnfsvolume.cloud-resources.kyma-project.io",
			"awsrediscluster.cloud-resources.kyma-project.io",
			"awsredisinstance.cloud-resources.kyma-project.io",
			"awsvpcpeering.cloud-resources.kyma-project.io",
			"iprange.cloud-resources.kyma-project.io",
		}
		expectedBusola := []string{
			"awsnfsbackupschedule.cloud-resources.kyma-project.io",
			"awsnfsvolumebackup.cloud-resources.kyma-project.io",
			"awsnfsvolumerestore.cloud-resources.kyma-project.io",
			"awsnfsvolume.cloud-resources.kyma-project.io",
			//"awsrediscluster.cloud-resources.kyma-project.io",
			"awsredisinstance.cloud-resources.kyma-project.io",
			"awsvpcpeering.cloud-resources.kyma-project.io",
			"iprange.cloud-resources.kyma-project.io",
		}

		run(context.Background(), t, cloudcontrolv1beta1.ProviderAws, expectedCrds, expectedBusola)
	})

	t.Run("azure", func(t *testing.T) {
		expectedCrds := []string{
			"azurerwxbackupschedule.cloud-resources.kyma-project.io",
			"azurerwxvolumebackup.cloud-resources.kyma-project.io",
			"azurerwxvolumerestore.cloud-resources.kyma-project.io",
			//"azurerwxvolume.cloud-resources.kyma-project.io",
			//"azurerediscluster.cloud-resources.kyma-project.io",
			"azureredisinstance.cloud-resources.kyma-project.io",
			"azurevpcpeering.cloud-resources.kyma-project.io",
			"iprange.cloud-resources.kyma-project.io",
		}
		expectedBusola := []string{
			"azurerwxbackupschedule.cloud-resources.kyma-project.io",
			//"azurerwxvolumebackup.cloud-resources.kyma-project.io",
			"azurerwxvolumerestore.cloud-resources.kyma-project.io",
			//"azurerwxvolume.cloud-resources.kyma-project.io",
			//"azurerediscluster.cloud-resources.kyma-project.io",
			"azureredisinstance.cloud-resources.kyma-project.io",
			"azurevpcpeering.cloud-resources.kyma-project.io",
			"iprange.cloud-resources.kyma-project.io",
		}

		run(context.Background(), t, cloudcontrolv1beta1.ProviderAzure, expectedCrds, expectedBusola)
	})

	t.Run("gcp", func(t *testing.T) {
		expectedCrds := []string{
			"gcpnfsbackupschedule.cloud-resources.kyma-project.io",
			"gcpnfsvolumebackup.cloud-resources.kyma-project.io",
			"gcpnfsvolumerestore.cloud-resources.kyma-project.io",
			"gcpnfsvolume.cloud-resources.kyma-project.io",
			//"gcprediscluster.cloud-resources.kyma-project.io",
			"gcpredisinstance.cloud-resources.kyma-project.io",
			"gcpvpcpeering.cloud-resources.kyma-project.io",
			"iprange.cloud-resources.kyma-project.io",
		}
		expectedBusola := []string{
			"gcpnfsbackupschedule.cloud-resources.kyma-project.io",
			"gcpnfsvolumebackup.cloud-resources.kyma-project.io",
			"gcpnfsvolumerestore.cloud-resources.kyma-project.io",
			"gcpnfsvolume.cloud-resources.kyma-project.io",
			//"gcprediscluster.cloud-resources.kyma-project.io",
			"gcpredisinstance.cloud-resources.kyma-project.io",
			"gcpvpcpeering.cloud-resources.kyma-project.io",
			"iprange.cloud-resources.kyma-project.io",
		}

		run(context.Background(), t, cloudcontrolv1beta1.ProviderGCP, expectedCrds, expectedBusola)
	})

	t.Run("openstack", func(t *testing.T) {
		expectedCrds := []string{
			"cceenfsvolume.cloud-resources.kyma-project.io",
		}
		expectedBusola := []string{}

		run(context.Background(), t, cloudcontrolv1beta1.ProviderOpenStack, expectedCrds, expectedBusola)
	})
}
