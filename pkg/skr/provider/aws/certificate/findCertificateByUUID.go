package certificate

import (
	"context"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func findCertificateByUUID(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.certificateArn != "" {
		return nil, ctx
	}

	targetId := state.ObjAsCertificate().Status.Id
	if targetId == "" {
		return nil, ctx
	}

	logger.WithValues("id", targetId).Info("Searching certificate by UUID tag")

	summaries, err := state.awsClient.ListCertificates(ctx)
	if err != nil {
		return composed.NewStatusPatcherComposed(state.ObjAsCertificate()).
			MutateStatus(func(c *cloudresourcesv1beta1.Certificate) {
				c.SetStatusProviderError(err.Error())
			}).
			OnStatusChanged(composed.Log("Failed to list certificates")).
			OnFailure(composed.Log("Error patching Certificate with list error")).
			OnSuccess(composed.Requeue).
			Run(ctx, state.Cluster().K8sClient())
	}

	var matched []string

	for _, summary := range summaries {
		if summary.CertificateArn == nil {
			continue
		}

		tags, err := state.awsClient.ListTagsForCertificate(ctx, *summary.CertificateArn)
		if err != nil {
			return composed.NewStatusPatcherComposed(state.ObjAsCertificate()).
				MutateStatus(func(c *cloudresourcesv1beta1.Certificate) {
					c.SetStatusProviderError(err.Error())
				}).
				OnStatusChanged(composed.Log("Failed to retrieve certificate tags")).
				OnFailure(composed.Log("Error patching Certificate with ListTags error")).
				OnSuccess(composed.Requeue).
				Run(ctx, state.Cluster().K8sClient())
		}

		for _, tag := range tags {
			if tag.Key != nil && *tag.Key == "Name" &&
				tag.Value != nil && *tag.Value == targetId {
				matched = append(matched, *summary.CertificateArn)
				break
			}
		}
	}

	if len(matched) == 0 {
		logger.Info("Certificate not found in ACM by UUID tag, will proceed to import")
		return nil, ctx
	}

	if len(matched) > 1 {
		errMsg := fmt.Sprintf("found %d certificates with Name tag=%s", len(matched), targetId)
		return composed.NewStatusPatcherComposed(state.ObjAsCertificate()).
			MutateStatus(func(c *cloudresourcesv1beta1.Certificate) {
				c.SetStatusProviderError(errMsg)
			}).
			OnStatusChanged(composed.Log("Multiple certificates with same cloud-manager id tag - data inconsistency")).
			OnFailure(composed.Log("Error patching Certificate with duplicate certificate error")).
			OnSuccess(composed.Requeue).
			Run(ctx, state.Cluster().K8sClient())
	}

	state.certificateArn = matched[0]
	logger.WithValues("arn", state.certificateArn).Info("Certificate found by UUID tag")
	return nil, ctx
}
