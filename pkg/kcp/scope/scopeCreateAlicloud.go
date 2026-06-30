package scope

import (
	"context"

	"github.com/elliotchance/pie/v2"
	gardeneraliclouddapi "github.com/gardener/gardener-extension-provider-alicloud/pkg/apis/alicloud/v1alpha1"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/utils/ptr"
)

func scopeCreateAlicloud(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	infra := &gardeneraliclouddapi.InfrastructureConfig{}
	err := json.Unmarshal(state.shoot.Spec.Provider.InfrastructureConfig.Raw, infra)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error unmarshalling AliCloud InfrastructureConfig", composed.StopAndForget, ctx)
	}

	scope := &cloudcontrolv1beta1.Scope{
		Spec: cloudcontrolv1beta1.ScopeSpec{
			Scope: cloudcontrolv1beta1.ScopeInfo{
				Alicloud: &cloudcontrolv1beta1.AlicloudScope{
					AccountId:  state.credentialData["accessKeyID"],
					VpcNetwork: common.GardenerVpcName(state.shootNamespace, state.shootName),
					Network: cloudcontrolv1beta1.AlicloudNetwork{
						Nodes:    ptr.Deref(state.shoot.Spec.Networking.Nodes, ""),
						Pods:     ptr.Deref(state.shoot.Spec.Networking.Pods, ""),
						Services: ptr.Deref(state.shoot.Spec.Networking.Services, ""),
						VPC: cloudcontrolv1beta1.AlicloudVPC{
							Id:   ptr.Deref(infra.Networks.VPC.ID, ""),
							CIDR: ptr.Deref(infra.Networks.VPC.CIDR, ""),
						},
						Zones: pie.Map(infra.Networks.Zones, func(z gardeneraliclouddapi.Zone) cloudcontrolv1beta1.AlicloudZone {
							return cloudcontrolv1beta1.AlicloudZone{
								Name:    z.Name,
								Workers: z.Workers,
							}
						}),
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

	return nil, ctx
}
