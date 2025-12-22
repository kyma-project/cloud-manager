package azurevpcdnslink

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func createKcpAzureVNetLink(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.KcpAzureVNetLink != nil {
		return nil, nil
	}

	state.KcpAzureVNetLink = (&cloudcontrolv1beta1.AzureVNetLinkBuilder{}).
		WithName(state.ObjAsVNetLink().Status.Id).
		WithNamespace(state.KymaRef.Namespace).
		WithLabels(map[string]string{
			common.LabelKymaModule: common.FieldOwner,
		}).
		WithAnnotations(map[string]string{
			cloudcontrolv1beta1.LabelKymaName:        state.KymaRef.Name,
			cloudcontrolv1beta1.LabelRemoteName:      state.ObjAsVNetLink().Name,
			cloudcontrolv1beta1.LabelRemoteNamespace: state.ObjAsVNetLink().Namespace,
		}).
		WithScope(state.KymaRef.Name).
		WithRemoteVirtualPrivateLinkName(state.ObjAsVNetLink().Spec.RemoteLinkName).
		WithRemotePrivateDnsZone(state.ObjAsVNetLink().Spec.RemotePrivateDnsZone).
		WithRemoteDnsResolverRuleset(state.ObjAsVNetLink().Spec.RemoteDnsResolverRuleset).
		WithRemoteTenant(state.ObjAsVNetLink().Spec.RemoteTenant).
		Build()

	err := state.KcpCluster.K8sClient().Create(ctx, state.KcpAzureVNetLink)

	if err == nil {
		logger.Info("Created KCP AzureVNetLink", "id", state.ObjAsVNetLink().Status.Id)
		return nil, ctx
	}

	return composed.LogErrorAndReturn(err, "Error creating KCP AzureVNetLink", composed.StopWithRequeue, ctx)
}
