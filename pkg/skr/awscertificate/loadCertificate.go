package awscertificate

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func loadCertificate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	cert := state.ObjAsAwsCertificate()

	// Skip if no ARN yet (first import)
	if cert.Status.Arn == "" {
		logger.Info("No ARN in status, certificate will be imported")
		return nil, ctx
	}

	// Load certificate detail from ACM
	detail, err := state.awsClient.DescribeCertificate(ctx, cert.Status.Arn)
	if err != nil {
		// If certificate not found in ACM, clear ARN and reimport
		if isNotFoundError(err) {
			logger.Info("Certificate not found in ACM, will reimport")
			return nil, ctx
		}
		return composed.LogErrorAndReturn(err, "Error loading certificate from ACM", composed.StopWithRequeue, ctx)
	}

	state.certificateDetail = detail
	state.certificateArn = *detail.CertificateArn
	logger.Info("Certificate loaded from ACM successfully")

	return nil, ctx
}

func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	// Check if error contains "ResourceNotFoundException"
	return client.IgnoreNotFound(err) == nil
}
