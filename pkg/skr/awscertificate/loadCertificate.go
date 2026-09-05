package awscertificate

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/acm"
	acmtypes "github.com/aws/aws-sdk-go-v2/service/acm/types"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func loadCertificate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	cert := state.ObjAsAwsCertificate()

	var arn string

	// FAST PATH: Try to load by ARN from status
	if cert.Status.Arn != "" {
		logger.Info("Loading certificate by ARN from status")
		arn = cert.Status.Arn
	} else {
		// FALLBACK: Search by serial number from parsed certificate data
		if state.certificateData == nil {
			// Should not happen - parseCertificateData runs before this action
			return composed.NewStatusPatcherComposed(state.ObjAsAwsCertificate()).
				MutateStatus(func(c *cloudresourcesv1beta1.AwsCertificate) {
					c.SetStatusProviderError("certificateData missing")
				}).
				OnStatusChanged(composed.Log("Cannot search certificate without parsed data")).
				OnFailure(composed.Log("Error patching AwsCertificate with status certificateData missing")).
				OnSuccess(composed.Requeue).
				Run(ctx, state.Cluster().K8sClient())
		}

		serialNumber := state.certificateData.SerialFormatted
		logger.
			WithValues("serial", serialNumber).
			Info("No ARN in status, searching by serial number from certificate data")

		// Build search filter for serial number
		filter := &acmtypes.CertificateFilterStatementMemberFilter{
			Value: &acmtypes.CertificateFilterMemberX509AttributeFilter{
				Value: &acmtypes.X509AttributeFilterMemberSerialNumber{
					Value: serialNumber,
				},
			},
		}

		searchInput := &acm.SearchCertificatesInput{
			FilterStatement: filter,
			MaxResults:      new(int32(100)),
		}

		results, err := state.awsClient.SearchCertificates(ctx, searchInput)
		if err != nil {
			return composed.NewStatusPatcherComposed(state.ObjAsAwsCertificate()).
				MutateStatus(func(c *cloudresourcesv1beta1.AwsCertificate) {
					c.SetStatusProviderError(err.Error())
				}).
				OnStatusChanged(composed.Log("Failed to search certificates")).
				OnFailure(composed.Log("Error patching AwsCertificate with search error")).
				OnSuccess(composed.Requeue).
				Run(ctx, state.Cluster().K8sClient())
		}

		// Filter by TagCloudManagerName matching metadata.name
		arns := make([]string, 0)
		targetName := cert.Name // metadata.name of the AwsCertificate resource

		for _, searchResult := range results {
			if searchResult.CertificateArn == nil {
				continue
			}

			// Retrieve tags for this certificate
			tags, err := state.awsClient.ListTagsForCertificate(ctx, *searchResult.CertificateArn)
			if err != nil {
				return composed.NewStatusPatcherComposed(state.ObjAsAwsCertificate()).
					MutateStatus(func(c *cloudresourcesv1beta1.AwsCertificate) {
						c.SetStatusProviderError(err.Error())
					}).
					OnStatusChanged(composed.Log("Failed to retrieve certificate tags")).
					OnFailure(composed.Log("Error patching AwsCertificate with ListTags error")).
					OnSuccess(composed.Requeue).
					Run(ctx, state.Cluster().K8sClient())
			}

			// Check if TagCloudManagerName matches the AwsCertificate name
			for _, tag := range tags {
				if tag.Key != nil && *tag.Key == common.TagCloudManagerName &&
					tag.Value != nil && *tag.Value == targetName {
					arns = append(arns, *searchResult.CertificateArn)
					break
				}
			}
		}

		if len(arns) == 0 {
			logger.Info("Certificate not found in ACM, will proceed to import")
			return nil, ctx // certificateDetail stays nil, importCertificate will run
		}

		if len(arns) > 1 {
			errMsg := fmt.Sprintf("found %d certificates with tag %s=%s", len(arns), common.TagCloudManagerName, targetName)
			return composed.NewStatusPatcherComposed(state.ObjAsAwsCertificate()).
				MutateStatus(func(c *cloudresourcesv1beta1.AwsCertificate) {
					c.SetStatusProviderError(errMsg)
				}).
				OnStatusChanged(composed.Log("Multiple certificates with same cloud-manager name tag - data inconsistency")).
				OnFailure(composed.Log("Error patching AwsCertificate with duplicate certificate error")).
				OnSuccess(composed.Requeue).
				Run(ctx, state.Cluster().K8sClient())
		}

		arn = arns[0]
		logger.WithValues("arn", arn).Info("Certificate found by tag-based search")
	}

	// Load certificate detail from ACM
	detail, err := state.awsClient.DescribeCertificate(ctx, arn)
	if err != nil {
		if isNotFoundError(err) {
			logger.Info("Certificate not found in ACM (stale ARN), will reimport")
			return nil, ctx // certificateDetail stays nil, importCertificate will run
		}
		return composed.NewStatusPatcherComposed(state.ObjAsAwsCertificate()).
			MutateStatus(func(c *cloudresourcesv1beta1.AwsCertificate) {
				c.SetStatusProviderError(err.Error())
			}).
			OnStatusChanged(composed.Log("Error loading certificate from ACM")).
			OnFailure(composed.Log("Error patching AwsCertificate with describe error")).
			OnSuccess(composed.Requeue).
			Run(ctx, state.Cluster().K8sClient())
	}

	// Populate state for downstream actions (compareCertificateData, updateStatus)
	state.certificateDetail = detail
	state.certificateArn = *detail.CertificateArn // ALWAYS set ARN in state

	logger.
		WithValues("arn", state.certificateArn).
		Info("Certificate loaded from ACM successfully")

	return nil, composed.LoggerIntoCtx(ctx, logger)
}

func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	// Check if error contains "ResourceNotFoundException"
	return client.IgnoreNotFound(err) == nil
}
