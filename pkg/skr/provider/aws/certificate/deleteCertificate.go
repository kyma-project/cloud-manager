package certificate

import (
	"context"
	"strings"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
)

func deleteCertificate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	cert := state.ObjAsCertificate()

	if cert.Status.ProviderId == "" {
		logger.Info("No ARN in status, nothing to delete")
		return nil, ctx
	}

	logger.Info("Deleting certificate from ACM", "arn", cert.Status.ProviderId)

	err := state.awsClient.DeleteCertificate(ctx, cert.Status.ProviderId)
	if err != nil {
		if awsmeta.IsNotFound(err) {
			logger.Info("Certificate not found in ACM, already deleted")
			return nil, ctx
		}

		if isResourceInUseException(err) {
			logger.Error(err, "Certificate is in use by AWS services, cannot delete")
			return composed.NewStatusPatcherComposed(cert).
				MutateStatus(func(c *cloudresourcesv1beta1.Certificate) {
					c.SetStatusDeletionBlockedByDependents("Certificate is in use by AWS services. Detach from all services before deleting.")
				}).
				OnSuccess(composed.Requeue).
				Run(ctx, state.Cluster().K8sClient())
		}

		logger.Error(err, "Error deleting certificate from ACM")
		return composed.NewStatusPatcherComposed(cert).
			MutateStatus(func(c *cloudresourcesv1beta1.Certificate) {
				c.SetStatusProviderError(err.Error())
			}).
			OnSuccess(composed.Requeue).
			Run(ctx, state.Cluster().K8sClient())
	}

	if cert.State() == cloudresourcesv1beta1.ReasonDeletionBlockedByDependents {
		logger.Info("Certificate is no longer in use, clearing DeletionBlockedByDependents state")
		err, _ := composed.NewStatusPatcherComposed(cert).
			MutateStatus(func(c *cloudresourcesv1beta1.Certificate) {
				c.RemoveStatusDeletionBlockedByDependents()
			}).
			OnFailure(composed.Log("Failed to remove DeletionBlockedByDependents state")).
			Run(ctx, state.Cluster().K8sClient())
		if err != nil {
			return err, ctx
		}
	}

	logger.Info("Certificate deleted from ACM successfully")
	return nil, ctx
}

func isResourceInUseException(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "ResourceInUseException")
}
