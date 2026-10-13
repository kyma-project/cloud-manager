package certificate

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/acm"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func importCertificate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	cert := state.ObjAsCertificate()

	if !state.certificateNeedsUpdate {
		logger.Info("Certificate update not needed, skipping import")
		return nil, ctx
	}

	logger.Info("Importing certificate to ACM")

	input := &acm.ImportCertificateInput{
		Certificate: state.certificateData.Certificate,
		PrivateKey:  state.certificateData.PrivateKey,
	}

	if cert.Status.ProviderId != "" {
		input.CertificateArn = new(cert.Status.ProviderId)
		logger.WithValues("arn", cert.Status.ProviderId).Info("Updating existing certificate")
	} else {
		input.Tags = state.convertTags()
		logger.Info("Creating new certificate with tags")
	}

	if len(state.certificateData.CertificateChain) > 0 {
		input.CertificateChain = state.certificateData.CertificateChain
	}

	arn, err := state.awsClient.ImportCertificate(ctx, input)
	if err != nil {
		logger.Error(err, "Error importing certificate to ACM")
		return composed.NewStatusPatcherComposed(cert).
			MutateStatus(func(c *cloudresourcesv1beta1.Certificate) {
				c.SetStatusProviderError("Error importing certificate to ACM: " + err.Error())
			}).
			OnSuccess(composed.Requeue).
			Run(ctx, state.Cluster().K8sClient())
	}

	state.certificateArn = arn

	logger.
		WithValues("arn", arn).
		Info("Certificate imported successfully")

	return nil, ctx
}
