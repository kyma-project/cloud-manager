package scope

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/common"

	"github.com/elliotchance/pie/v2"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	sapmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/meta"
	"k8s.io/utils/ptr"
)

func scopeCreateOpenStack(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	var zones []string
	for _, worker := range state.shoot.Spec.Provider.Workers {
		zones = append(zones, worker.Zones...)
	}
	zones = pie.Unique(zones)
	zones = pie.Sort(zones)
	scope := &cloudcontrolv1beta1.Scope{
		Spec: cloudcontrolv1beta1.ScopeSpec{
			Scope: cloudcontrolv1beta1.ScopeInfo{
				OpenStack: &cloudcontrolv1beta1.OpenStackScope{
					VpcNetwork: common.GardenerVpcName(state.shootNamespace, state.shootName), // ???
					DomainName: state.credentialData["domainName"],
					TenantName: state.credentialData["tenantName"],
					Network: cloudcontrolv1beta1.OpenStackNetwork{
						Nodes:    ptr.Deref(state.shoot.Spec.Networking.Nodes, ""),
						Pods:     ptr.Deref(state.shoot.Spec.Networking.Pods, ""),
						Services: ptr.Deref(state.shoot.Spec.Networking.Services, ""),
						Zones:    zones,
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

	ctx = sapmeta.SetSapDomainProjectRegion(
		ctx,
		scope.Spec.Scope.OpenStack.DomainName,
		scope.Spec.Scope.OpenStack.TenantName,
		scope.Spec.Region,
	)

	return nil, ctx
}
