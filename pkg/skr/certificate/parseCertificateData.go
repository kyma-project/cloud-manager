package certificate

import (
	"context"
	"encoding/pem"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	certificatetypes "github.com/kyma-project/cloud-manager/pkg/skr/certificate/types"
)

func parseCertificateData(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	cert := state.ObjAsCertificate()

	tlsCrt, ok := state.secret.Data["tls.crt"]
	if !ok || len(tlsCrt) == 0 {
		logger.Error(fmt.Errorf("missing tls.crt"), "Secret missing required key tls.crt")
		return composed.NewStatusPatcherComposed(cert).
			MutateStatus(func(c *cloudresourcesv1beta1.Certificate) {
				c.SetStatusProviderError("Secret missing required key tls.crt")
			}).
			OnSuccess(composed.Requeue).
			Run(ctx, state.Cluster().K8sClient())
	}

	privateKey, ok := state.secret.Data["tls.key"]
	if !ok || len(privateKey) == 0 {
		logger.Error(fmt.Errorf("missing tls.key"), "Secret missing required key tls.key")
		return composed.NewStatusPatcherComposed(cert).
			MutateStatus(func(c *cloudresourcesv1beta1.Certificate) {
				c.SetStatusProviderError("Secret missing required key tls.key")
			}).
			OnSuccess(composed.Requeue).
			Run(ctx, state.Cluster().K8sClient())
	}

	// tls.crt may contain a full chain; only the first PEM block (leaf) is imported.
	// Any intermediates embedded in tls.crt are intentionally discarded — use ca.crt for the chain.
	leafBlock, _ := pem.Decode(tlsCrt)
	if leafBlock == nil {
		logger.Error(fmt.Errorf("failed to decode PEM"), "Failed to decode certificate PEM")
		return composed.NewStatusPatcherComposed(cert).
			MutateStatus(func(c *cloudresourcesv1beta1.Certificate) {
				c.SetStatusProviderError("Failed to decode certificate PEM")
			}).
			OnSuccess(composed.Requeue).
			Run(ctx, state.Cluster().K8sClient())
	}

	state.data = &certificatetypes.CertificateData{
		Certificate:      pem.EncodeToMemory(leafBlock),
		PrivateKey:       privateKey,
		CertificateChain: state.secret.Data["ca.crt"],
	}

	return nil, ctx
}
