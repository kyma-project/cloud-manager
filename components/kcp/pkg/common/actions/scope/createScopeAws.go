package scope

import (
	"context"
	"fmt"
	"github.com/elliotchance/pie/v2"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-resources/components/kcp/api/cloud-resources/v1beta1"
	awsgardener "github.com/kyma-project/cloud-resources/components/kcp/pkg/provider/aws/gardener"
	"github.com/kyma-project/cloud-resources/components/lib/composed"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/utils/pointer"
)

func createScopeAws(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(State)

	// calling STS with Gardener credentials to find AWS Account ID
	stsClient, err := state.AwsGardenProvider().Sts()(
		ctx,
		state.Shoot().Spec.Region,
		state.CredentialData()["accessKeyID"],
		state.CredentialData()["secretAccessKey"],
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
	err = json.Unmarshal(state.Shoot().Spec.Provider.InfrastructureConfig.Raw, infra)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error unmarshalling InfrastructureConfig", composed.StopAndForget, nil)
	}

	scope := &cloudresourcesv1beta1.Scope{
		Spec: cloudresourcesv1beta1.ScopeSpec{
			Scope: cloudresourcesv1beta1.ScopeInfo{
				Aws: &cloudresourcesv1beta1.AwsScope{
					AccountId:  pointer.StringDeref(callerIdentity.Account, ""),
					VpcNetwork: commonVpcName(state.ShootNamespace(), state.ShootName()),
					Network: cloudresourcesv1beta1.AwsNetwork{
						VPC: cloudresourcesv1beta1.AwsVPC{
							Id:   pointer.StringDeref(infra.Networks.VPC.ID, ""),
							CIDR: pointer.StringDeref(infra.Networks.VPC.CIDR, ""),
						},
						Zones: pie.Map(infra.Networks.Zones, func(z awsgardener.Zone) cloudresourcesv1beta1.AwsZone {
							return cloudresourcesv1beta1.AwsZone{
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

	state.SetScope(scope)

	return nil, nil
}
