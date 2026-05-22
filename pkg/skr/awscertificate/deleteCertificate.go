package awscertificate

import (
	"context"
	"strings"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func deleteCertificate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	cert := state.ObjAsAwsCertificate()

	// Skip if no ARN (certificate never imported)
	if cert.Status.Arn == "" {
		logger.Info("No ARN in status, nothing to delete")
		return nil, ctx
	}

	logger.Info("Deleting certificate from ACM", "arn", cert.Status.Arn)

	err := state.awsClient.DeleteCertificate(ctx, cert.Status.Arn)
	if err != nil {
		// If certificate is in use, set error condition and requeue
		if isResourceInUseException(err) {
			logger.Error(err, "Certificate is in use by AWS services, cannot delete")
			return composed.NewStatusPatcherComposed(cert).
				MutateStatus(func(c *cloudresourcesv1beta1.AwsCertificate) {
					c.SetStatusProviderError("Certificate is in use by AWS services. Detach from all services before deleting.")
				}).
				OnSuccess(composed.Requeue).
				Run(ctx, state.Cluster().K8sClient())
		}

		// If certificate not found, it's already deleted
		if isNotFoundError(err) {
			logger.Info("Certificate not found in ACM, already deleted")
			return nil, ctx
		}

		return composed.LogErrorAndReturn(err, "Error deleting certificate from ACM", composed.StopWithRequeue, ctx)
	}

	logger.Info("Certificate deleted from ACM successfully")
	return nil, ctx
}

func isResourceInUseException(err error) bool {
	if err == nil {
		return false
	}
	// Check if error contains "ResourceInUseException"
	return strings.Contains(err.Error(), "ResourceInUseException")
}
