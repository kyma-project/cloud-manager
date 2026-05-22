package security

import (
	"context"

	"cloud.google.com/go/securitycentermanagement/apiv1/securitycentermanagementpb"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func sccServicesDisable(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	updated := false
	for svcID, svc := range state.sccServices {
		if svc.IntendedEnablementState == securitycentermanagementpb.SecurityCenterService_DISABLED {
			continue
		}
		logger.Info("Disabling SCC service", "serviceId", svcID)
		_, err := state.gcpClient.UpdateSecurityCenterService(ctx,
			&securitycentermanagementpb.SecurityCenterService{
				Name:                    svc.Name,
				IntendedEnablementState: securitycentermanagementpb.SecurityCenterService_DISABLED,
			},
			&fieldmaskpb.FieldMask{Paths: []string{"intended_enablement_state"}},
		)
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error disabling SCC service "+svcID, composed.StopWithRequeue, ctx)
		}
		updated = true
	}

	if updated {
		return composed.StopWithRequeue, ctx
	}
	return nil, ctx
}
