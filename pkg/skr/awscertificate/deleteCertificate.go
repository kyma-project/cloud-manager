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
		// If certificate not found, it's already deleted
		if isNotFoundError(err) {
			logger.Info("Certificate not found in ACM, already deleted")
			return nil, ctx
		}

		// If certificate is in use, set DeleteWhileUsed condition and requeue
		if isResourceInUseException(err) {
			logger.Error(err, "Certificate is in use by AWS services, cannot delete")
			return composed.NewStatusPatcherComposed(cert).
				MutateStatus(func(c *cloudresourcesv1beta1.AwsCertificate) {
					c.SetStatusDeleteWhileUsed("Certificate is in use by AWS services. Detach from all services before deleting.")
				}).
				OnSuccess(composed.Requeue).
				Run(ctx, state.Cluster().K8sClient())
		}

		// For any other error, set provider error and requeue
		logger.Error(err, "Error deleting certificate from ACM")
		return composed.NewStatusPatcherComposed(cert).
			MutateStatus(func(c *cloudresourcesv1beta1.AwsCertificate) {
				c.SetStatusProviderError(err.Error())
			}).
			OnSuccess(composed.Requeue).
			Run(ctx, state.Cluster().K8sClient())
	}

	// If deletion succeeded and we had a DeleteWhileUsed condition, remove it
	// Following pattern from pkg/kcp/subscription/statusSaveOnDelete.go
	if cert.Status.State == cloudresourcesv1beta1.ReasonDeleteWhileUsed {
		logger.Info("Certificate is no longer in use, removing DeleteWhileUsed condition")
		return composed.NewStatusPatcherComposed(cert).
			MutateStatus(func(c *cloudresourcesv1beta1.AwsCertificate) {
				c.RemoveStatusDeleteWhileUsed()
			}).
			OnSuccess(composed.Continue).
			OnFailure(composed.Log("Failed to remove DeleteWhileUsed condition")).
			Run(ctx, state.Cluster().K8sClient())
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
