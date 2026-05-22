package security

import (
	"context"
	"fmt"

	"cloud.google.com/go/securitycentermanagement/apiv1/securitycentermanagementpb"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

var sccServiceIDs = []string{
	"security-health-analytics",
	"web-security-scanner",
	"event-threat-detection",
	"vm-threat-detection",
}

func sccServicesLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	project := state.Subscription().Status.SubscriptionInfo.Gcp.Project

	state.sccServices = make(map[string]*securitycentermanagementpb.SecurityCenterService, len(sccServiceIDs))
	for _, svcID := range sccServiceIDs {
		name := fmt.Sprintf("projects/%s/locations/global/securityCenterServices/%s", project, svcID)
		svc, err := state.gcpClient.GetSecurityCenterService(ctx, name)
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error loading SCC service "+svcID, composed.StopWithRequeue, ctx)
		}
		state.sccServices[svcID] = svc
	}

	return nil, ctx
}
