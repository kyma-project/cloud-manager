package awsnfsvolume

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func createKcpNfsInstance(ctx context.Context, st composed.State) (error, context.Context) {
	//state := st.(*State)
	//logger := composed.LoggerFromCtx(ctx)
	//
	//if composed.MarkedForDeletionPredicate(ctx, st) {
	//	// SKR IpRange is marked for deletion, do not create mirror in KCP
	//	return nil, nil
	//}
	//
	//if state.KcpNfsInstance != nil {
	//	// mirror IpRange in KCP is already created
	//	return nil, nil
	//}
	//
	//state.KcpNfsInstance = &cloudcontrolv1beta1.NfsInstance{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name:      uuid.NewString(),
	//		Namespace: state.KymaRef.Namespace,
	//		Labels: map[string]string{
	//			cloudcontrolv1beta1.LabelKymaName:        state.KymaRef.Name,
	//			cloudcontrolv1beta1.LabelRemoteName:      state.Name().Name,
	//			cloudcontrolv1beta1.LabelRemoteNamespace: state.Name().Namespace,
	//		},
	//	},
	//	Spec: cloudcontrolv1beta1.NfsInstanceSpec{
	//		RemoteRef: cloudcontrolv1beta1.RemoteRef{},
	//		IpRange:   cloudcontrolv1beta1.IpRangeRef{},
	//		Scope:     cloudcontrolv1beta1.ScopeRef{},
	//		Instance:  cloudcontrolv1beta1.NfsInstanceInfo{},
	//	},
	//}

	return nil, nil
}
