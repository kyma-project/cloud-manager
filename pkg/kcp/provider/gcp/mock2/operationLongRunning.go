package mock2

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"cloud.google.com/go/filestore/apiv1/filestorepb"
	"cloud.google.com/go/longrunning"
	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	"github.com/googleapis/gax-go/v2"
	"github.com/googleapis/gax-go/v2/apierror"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/protoadapt"
	"google.golang.org/protobuf/types/known/anypb"
)

/*
done: true
metadata:
  '@type': type.googleapis.com/google.cloud.common.OperationMetadata
  apiVersion: v1
  cancelRequested: false
  createTime: '2026-02-19T15:16:20.607255509Z'
  endTime: '2026-02-19T15:19:15.207802363Z'
  target: projects/my-project/locations/us-east1-c/instances/cm-f6001aaa-9a9e-4aa9-b67b-400a000800b7
  verb: create
name: projects/my-project/locations/us-east1-c/operations/operation-1700014000790-6aa0aa0000dcc-04dec372-0a74207c
response:
  '@type': type.googleapis.com/google.cloud.filestore.v1.Instance
  createTime: '2026-02-19T15:16:20.602902646Z'
  description: f6001aaa-9a9e-4aa9-b67b-400a000800b7
  fileShares:
  - capacityGb: '1024'
    name: vol1
  name: projects/my-project/locations/us-east1-c/instances/cm-f6001aaa-9a9e-4aa9-b67b-400a000800b7
  networks:
  - connectMode: PRIVATE_SERVICE_ACCESS
    ipAddresses:
    - 10.251.0.2
    modes:
    - MODE_IPV4
    network: projects/my-project/global/networks/shoot--spm-test01--pp-63a0ba
    reservedIpRange: 10.251.0.0/29
  performanceLimits:
    maxIops: '600'
    maxReadIops: '600'
    maxReadThroughputBps: '104857600'
    maxWriteIops: '1000'
    maxWriteThroughputBps: '104857600'
  satisfiesPzs: false
  state: READY
  tier: BASIC_HDD
*/

// base operationLongRunning ===================================================

type operationLongRunning struct {
	proto *longrunningpb.Operation
}

func (o *operationLongRunning) Name() string {
	return o.proto.Name
}

func (o *operationLongRunning) Done() bool {
	return o.proto.Done
}

func (o *operationLongRunning) Poll(ctx context.Context, resp protoadapt.MessageV1) error {
	if util.IsContextDone(ctx) {
		return ctx.Err()
	}
	if !o.Done() {
		return nil
	}
	switch r := o.proto.Result.(type) {
	case *longrunningpb.Operation_Error:
		err, _ := apierror.FromError(status.ErrorProto(r.Error))
		return err
	case *longrunningpb.Operation_Response:
		if resp == nil {
			return nil
		}
		respV2 := protoadapt.MessageV2Of(resp)
		return anypb.UnmarshalTo(r.Response, respV2, proto.UnmarshalOptions{AllowPartial: true, DiscardUnknown: true})
	default:
		return fmt.Errorf("unsupported result type %[1]T: %[1]v", r)
	}
}

var DefaultWaitInterval = util.Timing.T60000ms()

func (o *operationLongRunning) Wait(ctx context.Context, resp protoadapt.MessageV1) error {
	return o.WaitWithInterval(ctx, resp, DefaultWaitInterval)
}

func (o *operationLongRunning) WaitWithInterval(ctx context.Context, resp protoadapt.MessageV1, interval time.Duration) error {
	bo := gax.Backoff{
		Initial: util.Timing.T1000ms(),
		Max:     interval,
	}
	if bo.Max < bo.Initial {
		bo.Max = bo.Initial
	}
	for {
		if err := o.Poll(ctx, resp); err != nil {
			return err
		}
		if o.Done() {
			return nil
		}
		if err := gax.Sleep(ctx, bo.Pause()); err != nil {
			return err
		}
	}
}

func (o *operationLongRunning) Metadata(meta protoadapt.MessageV1) error {
	if m := o.proto.Metadata; m != nil {
		metav2 := protoadapt.MessageV2Of(meta)
		return anypb.UnmarshalTo(m, metav2, proto.UnmarshalOptions{AllowPartial: true, DiscardUnknown: true})
	}
	return longrunning.ErrNoMetadata
}

var _ gcpclient.BaseOperation = (*operationLongRunning)(nil)

// void operation ===============================================================

type voidOperationLongRunning struct {
	op *operationLongRunning
}

var _ gcpclient.VoidOperation = (*voidOperationLongRunning)(nil)

func (o *voidOperationLongRunning) OperationPB() *longrunningpb.Operation {
	return o.op.proto
}

func (o *voidOperationLongRunning) Name() string {
	return o.op.Name()
}

func (o *voidOperationLongRunning) Done() bool {
	return o.op.Done()
}

func (o *voidOperationLongRunning) Poll(ctx context.Context, _ ...gax.CallOption) error {
	return o.op.Poll(ctx, nil)
}

func (o *voidOperationLongRunning) Wait(ctx context.Context, _ ...gax.CallOption) error {
	return o.op.Wait(ctx, nil)
}

// result operation ============================================================

type resultOperationLongRunning[T protoadapt.MessageV1] struct {
	op *operationLongRunning
}

var _ gcpclient.ResultOperation[*filestorepb.Instance] = (*resultOperationLongRunning[*filestorepb.Instance])(nil)

func (o *resultOperationLongRunning[T]) OperationPB() *longrunningpb.Operation {
	return o.op.proto
}

func (o *resultOperationLongRunning[T]) Name() string {
	return o.op.Name()
}

func (o *resultOperationLongRunning[T]) Done() bool {
	return o.op.Done()
}

func (o *resultOperationLongRunning[T]) Poll(ctx context.Context, _ ...gax.CallOption) (T, error) {
	var resp T
	if err := o.op.Poll(ctx, resp); err != nil {
		return resp, err
	}
	if !o.Done() {
		var zero T
		return zero, nil
	}
	return resp, nil
}

func (o *resultOperationLongRunning[T]) Wait(ctx context.Context, _ ...gax.CallOption) (T, error) {
	var result T
	t := reflect.TypeOf(result)

	if t.Kind() == reflect.Pointer {
		// T is a pointer type (e.g., *Instance), create new instance of the element type
		result = reflect.New(t.Elem()).Interface().(T)
	} else {
		// For non-pointer types (though Interface methods typically require pointer receivers)
		result = reflect.New(t).Elem().Interface().(T)
	}

	if err := o.op.Wait(ctx, result); err != nil {
		var zero T
		return zero, err
	}

	return result, nil
}
