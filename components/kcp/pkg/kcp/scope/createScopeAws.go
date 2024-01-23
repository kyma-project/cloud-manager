package scope

import (
	"context"
	"fmt"
	"github.com/elliotchance/pie/v2"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/components/kcp/api/cloud-control/v1beta1"
	awsgardener "github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/aws/gardener"
	"github.com/kyma-project/cloud-manager/components/lib/composed"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/utils/pointer"
)

func createScopeAws(ctx context.Context, st composed.State) (error, context.Context) {
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
			nil)
	}
	callerIdentity, err := stsClient.GetCallerIdentity(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(
			fmt.Errorf("error getting caller identity: %w", err),
			"Error creating AWS scope",
			composed.StopWithRequeue,
			nil)
	}

	infra := &awsgardener.InfrastructureConfig{}
	err = json.Unmarshal(state.shoot.Spec.Provider.InfrastructureConfig.Raw, infra)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error unmarshalling InfrastructureConfig", composed.StopAndForget, nil)
	}

	scope := &cloudcontrolv1beta1.Scope{
		Spec: cloudcontrolv1beta1.ScopeSpec{
			Scope: cloudcontrolv1beta1.ScopeInfo{
				Aws: &cloudcontrolv1beta1.AwsScope{
					AccountId:  pointer.StringDeref(callerIdentity.Account, ""),
					VpcNetwork: commonVpcName(state.shootNamespace, state.shootName),
					Network: cloudcontrolv1beta1.AwsNetwork{
						VPC: cloudcontrolv1beta1.AwsVPC{
							Id:   pointer.StringDeref(infra.Networks.VPC.ID, ""),
							CIDR: pointer.StringDeref(infra.Networks.VPC.CIDR, ""),
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

	state.SetObj(scope)

	return nil, nil
}
