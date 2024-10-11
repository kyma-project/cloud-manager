package scope

import (
	"context"
	"errors"
	"fmt"
	"github.com/elliotchance/pie/v2"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuregardener "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/gardener"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/utils/ptr"
)

func createScopeAzure(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(*State)

	subscriptionID, ok := state.credentialData["subscriptionID"]
	if !ok {
		err := errors.New("gardener credential for azure missing subscriptionID key")
		logger.Error(err, "error defining Azure scope")
		return composed.StopAndForget, nil // no requeue
	}

	tenantID, ok := state.credentialData["tenantID"]
	if !ok {
		err := errors.New("gardener credential for azure missing tenantID key")
		logger.Error(err, "error defining Azure scope")
		return composed.StopAndForget, nil // no requeue
	}

	infra := &azuregardener.InfrastructureConfig{}
	err := json.Unmarshal(state.shoot.Spec.Provider.InfrastructureConfig.Raw, infra)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error unmarshalling Azure InfrastructureConfig", composed.StopAndForget, ctx)
	}

	// just create the scope with Azure specifics, the ensureScopeCommonFields will set common values
	scope := &cloudcontrolv1beta1.Scope{
		Spec: cloudcontrolv1beta1.ScopeSpec{
			Scope: cloudcontrolv1beta1.ScopeInfo{
				Azure: &cloudcontrolv1beta1.AzureScope{
					TenantId:       tenantID,
					SubscriptionId: subscriptionID,
					VpcNetwork:     commonVpcName(state.shootNamespace, state.shootName),
					Network: cloudcontrolv1beta1.AzureNetwork{
						Cidr: ptr.Deref(infra.Networks.VNet.CIDR, ""),
						Zones: pie.Map(infra.Networks.Zones, func(z azuregardener.Zone) cloudcontrolv1beta1.AzureNetworkZone {
							return cloudcontrolv1beta1.AzureNetworkZone{
								Name: fmt.Sprintf("%d", z.Name),
								Cidr: z.CIDR,
							}
						}),
						Nodes:    ptr.Deref(state.shoot.Spec.Networking.Nodes, ""),
						Pods:     ptr.Deref(state.shoot.Spec.Networking.Pods, ""),
						Services: ptr.Deref(state.shoot.Spec.Networking.Services, ""),
					},
				},
			},
		},
	}

	// Preserve loaded obj resource version before getting overwritten by newly created scope
	if st.Obj() != nil && st.Obj().GetName() != "" {
		scope.ResourceVersion = st.Obj().GetResourceVersion()
	}
	state.SetObj(scope)

	return nil, nil
}
