package scope

import (
	"context"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createKymaNetworkReference(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	kymaNet := state.allNetworks.FindFirstByType(cloudcontrolv1beta1.NetworkTypeKyma)
	if kymaNet != nil {
		return nil, nil
	}

	kymaNet = &cloudcontrolv1beta1.Network{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s--kyma", state.ObjAsScope().Name),
			Namespace: state.ObjAsScope().Namespace,
		},
		Spec: cloudcontrolv1beta1.NetworkSpec{
			Type: cloudcontrolv1beta1.NetworkTypeKyma,
			Scope: cloudcontrolv1beta1.ScopeRef{
				Name: state.ObjAsScope().Name,
			},
		},
	}

	switch state.provider {
	case cloudcontrolv1beta1.ProviderGCP:
		kymaNet.Spec.Network.Reference = &cloudcontrolv1beta1.NetworkReference{
			Gcp: &cloudcontrolv1beta1.GcpNetworkReference{
				GcpProject:  state.ObjAsScope().Spec.Scope.Gcp.Project,
				NetworkName: state.ObjAsScope().Spec.Scope.Gcp.VpcNetwork,
			},
		}
	case cloudcontrolv1beta1.ProviderAzure:
		kymaNet.Spec.Network.Reference = &cloudcontrolv1beta1.NetworkReference{
			Azure: &cloudcontrolv1beta1.AzureNetworkReference{
				TenantId:       state.ObjAsScope().Spec.Scope.Azure.TenantId,
				SubscriptionId: state.ObjAsScope().Spec.Scope.Azure.SubscriptionId,
				ResourceGroup:  state.ObjAsScope().Spec.Scope.Azure.VpcNetwork,
				NetworkName:    state.ObjAsScope().Spec.Scope.Azure.VpcNetwork,
			},
		}
	case cloudcontrolv1beta1.ProviderAws:
		kymaNet.Spec.Network.Reference = &cloudcontrolv1beta1.NetworkReference{
			Aws: &cloudcontrolv1beta1.AwsNetworkReference{
				AwsAccountId: state.ObjAsScope().Spec.Scope.Aws.AccountId,
				Region:       state.ObjAsScope().Spec.Region,
				NetworkName:  state.ObjAsScope().Spec.Scope.Aws.VpcNetwork,
			},
		}
	case cloudcontrolv1beta1.ProviderOpenStack:
		kymaNet.Spec.Network.Reference = &cloudcontrolv1beta1.NetworkReference{
			OpenStack: &cloudcontrolv1beta1.OpenStackNetworkReference{
				Domain:      state.ObjAsScope().Spec.Scope.OpenStack.DomainName,
				Project:     state.ObjAsScope().Spec.Scope.OpenStack.TenantName,
				NetworkName: state.ObjAsScope().Spec.Scope.OpenStack.VpcNetwork,
			},
		}
	}

	err := state.Cluster().K8sClient().Create(ctx, kymaNet)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error saving kyma network", composed.StopWithRequeue, ctx)
	}

	return nil, nil
}
