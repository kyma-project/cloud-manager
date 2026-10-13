package certificate

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
)

func loadCertificate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.certificateArn == "" {
		return nil, ctx
	}

	detail, err := state.awsClient.DescribeCertificate(ctx, state.certificateArn)
	if err != nil {
		if awsmeta.IsNotFound(err) {
			logger.Info("Certificate not found in ACM (stale ARN), will reimport")
			state.certificateArn = ""
			return nil, ctx
		}
		return composed.NewStatusPatcherComposed(state.ObjAsCertificate()).
			MutateStatus(func(c *cloudresourcesv1beta1.Certificate) {
				c.SetStatusProviderError(err.Error())
			}).
			OnStatusChanged(composed.Log("Error loading certificate from ACM")).
			OnSuccess(composed.Requeue).
			Run(ctx, state.Cluster().K8sClient())
	}

	state.certificateDetail = detail
	state.certificateArn = *detail.CertificateArn

	logger.WithValues("arn", state.certificateArn).Info("Certificate loaded from ACM")
	return nil, ctx
}
