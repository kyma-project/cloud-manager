package mock2

import (
	"context"
	"fmt"
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

type operation struct {
	proto *longrunningpb.Operation
}

func (o *operation) Name() string {
	return o.proto.Name
}

func (o *operation) Done() bool {
	return o.proto.Done
}

func (o *operation) Poll(ctx context.Context, resp protoadapt.MessageV1) error {
	if util.IsContextDone(ctx) {
		return ctx.Err()
	}
	if !o.proto.Done {
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
		respv2 := protoadapt.MessageV2Of(resp)
		return anypb.UnmarshalTo(r.Response, respv2, proto.UnmarshalOptions{AllowPartial: true, DiscardUnknown: true})
	default:
		return fmt.Errorf("unsupported result type %[1]T: %[1]v", r)
	}
}

const DefaultWaitInterval = longrunning.DefaultWaitInterval

func (o *operation) Wait(ctx context.Context, resp protoadapt.MessageV1) error {
	return o.WaitWithInterval(ctx, resp, DefaultWaitInterval)
}

func (o *operation) WaitWithInterval(ctx context.Context, resp protoadapt.MessageV1, interval time.Duration) error {
	bo := gax.Backoff{
		Initial: 1 * time.Second,
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

var _ gcpclient.BaseOperation = (*operation)(nil)

// void operation ===============================================================

type voidOperation struct {
	op *operation
}

var _ gcpclient.VoidOperation = (*voidOperation)(nil)

func (o *voidOperation) Name() string {
	return o.Name()
}

func (o *voidOperation) Done() bool {
	return o.Done()
}

func (o *voidOperation) Poll(ctx context.Context, _ ...gax.CallOption) error {
	return o.op.Poll(ctx, nil)
}

func (o *voidOperation) Wait(ctx context.Context, _ ...gax.CallOption) error {
	return o.op.Wait(ctx, nil)
}

// result operation ============================================================

type resultOperation[T protoadapt.MessageV1] struct {
	op *operation
}

var _ gcpclient.ResultOperation[*filestorepb.Instance] = (*resultOperation[*filestorepb.Instance])(nil)

func (o *resultOperation[T]) Name() string {
	return o.op.Name()
}

func (o *resultOperation[T]) Done() bool {
	return o.op.Done()
}

func (o *resultOperation[T]) Poll(ctx context.Context, opts ...gax.CallOption) (T, error) {
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

func (o *resultOperation[T]) Wait(ctx context.Context, opts ...gax.CallOption) (T, error) {
	var resp T
	if err := o.op.Wait(ctx, resp); err != nil {
		var zero T
		return zero, err
	}
	return resp, nil
}


//
//func newVoidOperation(name string, done bool, err error) *voidOperation {
//	return &voidOperation{
//		operation: operation{
//			name: name,
//			done: done,
//		},
//		err: err,
//	}
//}
//
//func newVoidOperationFromComputeOperation(op *computepb.Operation) *voidOperation {
//	done := ptr.Deref(op.Status, computepb.Operation_UNDEFINED_STATUS) == computepb.Operation_DONE
//	var err error
//	if op.Error != nil {
//		err = errors.New(op.Error.String())
//	}
//	return newVoidOperation(ptr.Deref(op.Name, ""), done, err)
//}
//
//var _ gcpclient.VoidOperation = (*voidOperation)(nil)
//
//func (o *voidOperation) Wait(ctx context.Context, opts ...gax.CallOption) error {
//	for {
//		if util.IsContextDone(ctx) {
//			return ctx.Err()
//		}
//		if o.Done() {
//			return o.err
//		}
//		time.Sleep(10 * time.Millisecond)
//	}
//}
//
//func (o *voidOperation) Poll(ctx context.Context, opts ...gax.CallOption) error {
//
//}
//
//type operationWithResult[T any] struct {
//	operation
//	result T
//	err    error
//}
//
//func newOperationWithResult[T any](name string, done bool, result T, err error) *operationWithResult[T] {
//	return &operationWithResult[T]{
//		operation: operation{
//			name: name,
//			done: done,
//		},
//		result: result,
//		err:    err,
//	}
//}
//
//var _ gcpclient.WaitableOperationWithResult[any] = (*operationWithResult[any])(nil)
//
//func (o operationWithResult[T]) Wait(ctx context.Context, opts ...gax.CallOption) (T, error) {
//	time.Sleep(time.Millisecond)
//	return o.result, o.err
//}
