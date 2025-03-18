package k8sport

import (
	"context"
	composedv2 "github.com/kyma-project/cloud-manager/pkg/composed/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type K8sEventPort interface {
	// Event constructs an event from the given information and puts it in the queue for sending.
	// 'object' is the object this event is about. Event will make a reference-- or you may also
	// pass a reference to the object directly.
	// 'eventtype' of this event, and can be one of Normal, Warning. New types could be added in future
	// 'reason' is the reason this event is generated. 'reason' should be short and unique; it
	// should be in UpperCamelCase format (starting with a capital letter). "reason" will be used
	// to automate handling of events, so imagine people writing switch statements to handle them.
	// You want to make that easy.
	// 'message' is intended to be human readable.
	//
	// The resulting event will be created in the same namespace as the reference object.
	Event(ctx context.Context, object client.Object, eventtype, reason, message string)
	Eventf(ctx context.Context, object client.Object, eventtype, reason, messageFmt string, args ...interface{})
	AnnotatedEventf(ctx context.Context, object client.Object, annotations map[string]string, eventtype, reason, messageFmt string, args ...interface{})
}

func NewK8sEventPort(clusterID string) K8sEventPort {
	return &k8sEventPortImpl{clusterID: clusterID}
}

func NewK8sEventPortOnDefaultCluster() K8sEventPort {
	return NewK8sEventPort(composedv2.DefaultClusterID)
}

type k8sEventPortImpl struct {
	clusterID string
}

func (p *k8sEventPortImpl) Event(ctx context.Context, object client.Object, eventtype, reason, message string) {
	cluster := composedv2.ClusterFromCtx(ctx, p.clusterID)
	cluster.EventRecorder().Event(object, eventtype, reason, message)
}

func (p *k8sEventPortImpl) Eventf(ctx context.Context, object client.Object, eventtype, reason, messageFmt string, args ...interface{}) {
	cluster := composedv2.ClusterFromCtx(ctx, p.clusterID)
	cluster.EventRecorder().Eventf(object, eventtype, reason, messageFmt, args...)
}

func (p *k8sEventPortImpl) AnnotatedEventf(ctx context.Context, object client.Object, annotations map[string]string, eventtype, reason, messageFmt string, args ...interface{}) {
	cluster := composedv2.ClusterFromCtx(ctx, p.clusterID)
	cluster.EventRecorder().AnnotatedEventf(object, annotations, eventtype, reason, messageFmt, args...)
}
