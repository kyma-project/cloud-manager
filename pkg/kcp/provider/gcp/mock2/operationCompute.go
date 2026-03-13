package mock2

import (
	"context"
	"fmt"

	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/googleapis/gax-go/v2"
	"github.com/googleapis/gax-go/v2/apierror"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"google.golang.org/api/googleapi"
)

/*
{
  "endTime": "2026-02-19T23:02:58.025-08:00",
  "id": "6378747150325889486",
  "insertTime": "2026-02-19T23:02:57.902-08:00",
  "kind": "compute#operation",
  "name": "operation-1771570977787-64b3c02d34d2f-e2fad38f-7444a00d",
  "operationType": "patch",
  "progress": 100,
  "region": "https://www.googleapis.com/compute/v1/projects/my-gcp-project/regions/us-east1",
  "selfLink": "https://www.googleapis.com/compute/v1/projects/my-gcp-project/regions/us-east1/operations/operation-1771570977787-64b3c02d34d2f-e2fad38f-7444a00d",
  "startTime": "2026-02-19T23:02:57.913-08:00",
  "status": "DONE",
  "targetId": "3075952161956352499",
  "targetLink": "https://www.googleapis.com/compute/v1/projects/my-gcp-project/regions/us-east1/routers/my-router",
  "user": "gardener-test@my-gcp-project.iam.gserviceaccount.com"
}
*/

// operationCompute ============================================

type operationCompute struct {
	proto *computepb.Operation
}

var _ gcpclient.VoidOperation = (*operationCompute)(nil)

func (o *operationCompute) Name() string {
	return o.proto.GetName()
}

func (o *operationCompute) Done() bool {
	return o.proto.GetStatus() == computepb.Operation_DONE
}

func (o *operationCompute) Poll(ctx context.Context, _ ...gax.CallOption) error {
	if util.IsContextDone(ctx) {
		return ctx.Err()
	}
	if !o.Done() {
		return nil
	}
	if o.proto.HttpErrorStatusCode != nil && (o.proto.GetHttpErrorStatusCode() < 200 || o.proto.GetHttpErrorStatusCode() > 299) {
		aErr := &googleapi.Error{
			Code:    int(o.proto.GetHttpErrorStatusCode()),
			Message: fmt.Sprintf("%s: %v", o.proto.GetHttpErrorMessage(), o.proto.GetError()),
		}
		err, _ := apierror.FromError(aErr)
		return err
	}
	return nil
}

func (o *operationCompute) Wait(ctx context.Context, opts ...gax.CallOption) error {
	bo := gax.Backoff{
		Initial: util.Timing.T1000ms(),
		Max:     DefaultWaitInterval,
	}
	for {
		if err := o.Poll(ctx, opts...); err != nil {
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

func newComputeOperation(proto *computepb.Operation) *operationCompute {
	return &operationCompute{proto: proto}
}
