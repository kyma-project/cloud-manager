package scope

import (
	"context"
	"fmt"
	"github.com/elliotchance/pie/v2"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsgardener "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/gardener"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/utils/ptr"
)

func scopeCreateAws(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	// calling STS with Gardener credentials to find AWS Account ID
	stsClient, err := state.awsStsClientProvider(
		ctx,
		state.shoot.Spec.Region,
		state.credentialData["accessKeyID"],
		state.credentialData["secretAccessKey"],
	)
	if err != nil {
		return composed.LogErrorAndReturn(
			fmt.Errorf("error creating aws scope: %w", err),
			"Error creating AWS scope",
			composed.StopAndForget,
			ctx)
	}
	callerIdentity, err := stsClient.GetCallerIdentity(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(
			fmt.Errorf("error getting caller identity: %w", err),
			"Error creating AWS scope",
			composed.StopWithRequeue,
			ctx)
	}

	infra := &awsgardener.InfrastructureConfig{}
	err = json.Unmarshal(state.shoot.Spec.Provider.InfrastructureConfig.Raw, infra)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error unmarshalling AWS InfrastructureConfig", composed.StopAndForget, ctx)
	}

	scope := &cloudcontrolv1beta1.Scope{
		Spec: cloudcontrolv1beta1.ScopeSpec{
			Scope: cloudcontrolv1beta1.ScopeInfo{
				Aws: &cloudcontrolv1beta1.AwsScope{
					AccountId:  ptr.Deref(callerIdentity.Account, ""),
					VpcNetwork: common.GardenerVpcName(state.shootNamespace, state.shootName),
					Network: cloudcontrolv1beta1.AwsNetwork{
						Nodes:    ptr.Deref(state.shoot.Spec.Networking.Nodes, ""),
						Pods:     ptr.Deref(state.shoot.Spec.Networking.Pods, ""),
						Services: ptr.Deref(state.shoot.Spec.Networking.Services, ""),
						VPC: cloudcontrolv1beta1.AwsVPC{
							Id:   ptr.Deref(infra.Networks.VPC.ID, ""),
							CIDR: ptr.Deref(infra.Networks.VPC.CIDR, ""),
						},
						Zones: pie.Map(infra.Networks.Zones, func(z awsgardener.Zone) cloudcontrolv1beta1.AwsZone {
							return cloudcontrolv1beta1.AwsZone{
								Name:     z.Name,
								Internal: z.Internal,
								Public:   z.Public,
								Workers:  z.Workers,
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
