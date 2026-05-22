package awscertificate

import (
	"context"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func parseCertificateData(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	cert := state.ObjAsAwsCertificate()

	// Extract certificate data from Secret
	certificate, ok := state.secret.Data["tls.crt"]
	if !ok || len(certificate) == 0 {
		logger.Error(fmt.Errorf("missing tls.crt"), "Secret missing required key tls.crt")
		return composed.NewStatusPatcherComposed(cert).
			MutateStatus(func(c *cloudresourcesv1beta1.AwsCertificate) {
				c.SetStatusProviderError("Secret missing required key tls.crt")
			}).
			OnSuccess(composed.Requeue).
			Run(ctx, state.Cluster().K8sClient())
	}

	privateKey, ok := state.secret.Data["tls.key"]
	if !ok || len(privateKey) == 0 {
		logger.Error(fmt.Errorf("missing tls.key"), "Secret missing required key tls.key")
		return composed.NewStatusPatcherComposed(cert).
			MutateStatus(func(c *cloudresourcesv1beta1.AwsCertificate) {
				c.SetStatusProviderError("Secret missing required key tls.key")
			}).
			OnSuccess(composed.Requeue).
			Run(ctx, state.Cluster().K8sClient())
	}

	// Certificate chain is optional
	certificateChain := state.secret.Data["ca.crt"]

	state.certificateData = &CertificateData{
		Certificate:      certificate,
		PrivateKey:       privateKey,
		CertificateChain: certificateChain,
	}

	logger.Info("Certificate data parsed successfully")
	return nil, ctx
}
