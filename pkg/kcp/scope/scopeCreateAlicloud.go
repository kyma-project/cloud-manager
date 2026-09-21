package scope

import (
	"context"
	"fmt"

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

	// Call STS with the shoot (runtime account) credentials to find the AliCloud
	// account id, needed to build the assume-role ARN acs:ram::<accountId>:role/CloudManagerRole.
	stsClient, err := state.alicloudStsClientProvider(
		ctx,
		state.shoot.Spec.Region,
		state.credentialData["accessKeyID"],
		state.credentialData["accessKeySecret"],
	)
	if err != nil {
		return composed.LogErrorAndReturn(
			fmt.Errorf("error creating alicloud sts client: %w", err),
			"Error creating AliCloud scope",
			composed.StopAndForget,
			ctx)
	}
	accountId, err := stsClient.GetCallerIdentity(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(
			fmt.Errorf("error getting caller identity: %w", err),
			"Error creating AliCloud scope",
			composed.StopWithRequeue,
			ctx)
	}

	infra := &gardeneraliclouddapi.InfrastructureConfig{}
	err = json.Unmarshal(state.shoot.Spec.Provider.InfrastructureConfig.Raw, infra)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error unmarshalling AliCloud InfrastructureConfig", composed.StopAndForget, ctx)
	}

	scope := &cloudcontrolv1beta1.Scope{
		Spec: cloudcontrolv1beta1.ScopeSpec{
			Scope: cloudcontrolv1beta1.ScopeInfo{
				Alicloud: &cloudcontrolv1beta1.AlicloudScope{
					AccountId:  accountId,
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
