package mock2

import (
	"fmt"
	"reflect"

	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	"cloud.google.com/go/networkconnectivity/apiv1/networkconnectivitypb"
	"cloud.google.com/go/redis/apiv1/redispb"
	"cloud.google.com/go/redis/cluster/apiv1/clusterpb"
	"github.com/kyma-project/cloud-manager/pkg/common"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcputil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/util"
	commonpb "google.golang.org/genproto/googleapis/cloud/common"
	rpcstatus "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/protoadapt"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OperationLongRunningBuilder struct {
	proto           *longrunningpb.Operation
	relatedItemName gcputil.NameDetail
}

func NewOperationLongRunningBuilder(operationName string, relatedItemName gcputil.NameDetail) *OperationLongRunningBuilder {
	return &OperationLongRunningBuilder{
		proto: &longrunningpb.Operation{
			Name: operationName,
		},
		relatedItemName: relatedItemName,
	}
}

func (b *OperationLongRunningBuilder) RelatedItemName() gcputil.NameDetail {
	return b.relatedItemName
}

func (b *OperationLongRunningBuilder) WithOperation(op *longrunningpb.Operation) *OperationLongRunningBuilder {
	b.proto = op
	return b
}

func (b *OperationLongRunningBuilder) WithDone(done bool) *OperationLongRunningBuilder {
	b.proto.Done = done
	return b
}

func (b *OperationLongRunningBuilder) WithSimpleError(code int32, message string) *OperationLongRunningBuilder {
	b.proto.Result = &longrunningpb.Operation_Error{
		Error: &rpcstatus.Status{
			Code:    code,
			Message: message,
		},
	}
	return b
}

func (b *OperationLongRunningBuilder) WithOperationError(err *longrunningpb.Operation_Error) *OperationLongRunningBuilder {
	b.proto.Result = err
	return b
}

func (b *OperationLongRunningBuilder) WithResult(result protoadapt.MessageV1) error {
	dst := &anypb.Any{}
	respV2 := protoadapt.MessageV2Of(result)
	if err := anypb.MarshalFrom(dst, respV2, proto.MarshalOptions{}); err != nil {
		return fmt.Errorf("%w: failed to marshal operation result into anypb: %w", common.ErrLogical, err)
	}
	b.proto.Result = &longrunningpb.Operation_Response{
		Response: dst,
	}
	return nil
}

func (b *OperationLongRunningBuilder) WithMetadata(meta protoadapt.MessageV1) error {
	dst := &anypb.Any{}
	respV2 := protoadapt.MessageV2Of(meta)
	if err := anypb.MarshalFrom(dst, respV2, proto.MarshalOptions{}); err != nil {
		return fmt.Errorf("%w: failed to marshal operation metadata into anypb: %w", common.ErrLogical, err)
	}
	b.proto.Metadata = dst
	return nil
}

func (b *OperationLongRunningBuilder) WithCommonMetadata(targetName gcputil.NameDetail, verb string) error {
	return b.WithMetadata(&commonpb.OperationMetadata{
		CreateTime: timestamppb.Now(),
		Target:     targetName.String(),
		Verb:       verb,
		ApiVersion: "v1",
	})
}

func (b *OperationLongRunningBuilder) WithRedisInstanceMetadata(targetName gcputil.NameDetail, verb string) error {
	return b.WithMetadata(&redispb.OperationMetadata{
		CreateTime: timestamppb.Now(),
		Target:     targetName.String(),
		Verb:       verb,
		ApiVersion: "v1",
	})
}

func (b *OperationLongRunningBuilder) WithRedisClusterMetadata(targetName gcputil.NameDetail, verb string) error {
	return b.WithMetadata(&clusterpb.OperationMetadata{
		CreateTime: timestamppb.Now(),
		Target:     targetName.String(),
		Verb:       verb,
		ApiVersion: "v1",
	})
}

func (b *OperationLongRunningBuilder) WithNetworkConnectivityMetadata(targetName gcputil.NameDetail, verb string) error {
	return b.WithMetadata(&networkconnectivitypb.OperationMetadata{
		CreateTime: timestamppb.Now(),
		Target:     targetName.String(),
		Verb:       verb,
		ApiVersion: "v1",
	})
}

func (b *OperationLongRunningBuilder) GetOperationPB() *longrunningpb.Operation {
	return b.proto
}

func (b *OperationLongRunningBuilder) BuildVoidOperation() gcpclient.VoidOperation {
	return &voidOperationLongRunning{
		op: &operationLongRunning{
			proto: b.proto,
		},
	}
}

func NewResultOperation[T protoadapt.MessageV1](pb *longrunningpb.Operation) gcpclient.ResultOperation[T] {
	return &resultOperationLongRunning[T]{
		op: &operationLongRunning{
			proto: pb,
		},
	}
}

var _ protoadapt.MessageV2 = (*commonpb.OperationMetadata)(nil)

func ReadOperationMetadata[T protoadapt.MessageV1](b *OperationLongRunningBuilder) (T, error) {
	var meta T
	t := reflect.TypeOf(meta)

	if t.Kind() == reflect.Ptr {
		// T is a pointer type (e.g., *Instance), create new instance of the element type
		meta = reflect.New(t.Elem()).Interface().(T)
	} else {
		// For non-pointer types (though Interface methods typically require pointer receivers)
		meta = reflect.New(t).Elem().Interface().(T)
	}

	var zero T

	if m := b.proto.Metadata; m != nil {
		metav2 := protoadapt.MessageV2Of(meta)
		err := anypb.UnmarshalTo(m, metav2, proto.UnmarshalOptions{AllowPartial: true, DiscardUnknown: true})
		if err != nil {
			return zero, err
		}
		return metav2.(T), nil
	}
	return zero, nil
}
