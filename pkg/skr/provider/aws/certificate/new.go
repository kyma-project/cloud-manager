package certificate

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	certificatetypes "github.com/kyma-project/cloud-manager/pkg/skr/certificate/types"
)

func New(stateFactory StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		logger := composed.LoggerFromCtx(ctx)
		certState := st.(certificatetypes.State)
		state, err := stateFactory.NewState(ctx, certState)
		if err != nil {
			err = fmt.Errorf("error creating new aws certificate state: %w", err)
			logger.Error(err, "Error")
			return composed.StopAndForget, ctx
		}

		return composed.ComposeActions(
			"awsCertificate",
			createAwsClient,
			loadSecret,
			parseCertificateData,
			// create/update =========================================================================
			composed.If(composed.NotMarkedForDeletionPredicate,
				composed.ComposeActionsNoName(
					actions.AddCommonFinalizer(),
					updateId,
					findCertificateByProviderId,
					findCertificateByUUID,
					loadCertificate,
					compareCertificateData,
					statusProcessing,
					importCertificate,
					updateStatus,
				),
			),
			// delete ================================================================================
			composed.If(composed.MarkedForDeletionPredicate,
				composed.ComposeActionsNoName(
					statusDeleting,
					deleteCertificate,
					actions.RemoveCommonFinalizer(),
					composed.StopAndForgetAction,
				),
			),
			composed.StopAndForgetAction,
		)(ctx, state)
	}
}
