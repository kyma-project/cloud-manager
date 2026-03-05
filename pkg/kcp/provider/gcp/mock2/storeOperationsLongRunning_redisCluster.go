package mock2

import (
	"context"

	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	"github.com/googleapis/gax-go/v2"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

// Client Interface implementations ===================================================

func (s *store) GetRedisClusterOperation(ctx context.Context, req *longrunningpb.GetOperationRequest, _ ...gax.CallOption) (*longrunningpb.Operation, error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	return s.getLongRunningOperationNoLock(req.Name)
}

func (s *store) ListRedisClusterOperations(ctx context.Context, req *longrunningpb.ListOperationsRequest, _ ...gax.CallOption) gcpclient.Iterator[*longrunningpb.Operation] {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return &iteratorMocked[*longrunningpb.Operation]{
			err: ctx.Err(),
		}
	}

	return s.listLongRunningOperationsNoLock(req.Name, req.Filter)
}
