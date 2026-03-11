package mock2

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	"cloud.google.com/go/networkconnectivity/apiv1/networkconnectivitypb"
	"github.com/googleapis/gax-go/v2"
	"github.com/kyma-project/cloud-manager/pkg/common"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
	gcputil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

/*
done: true
metadata:
  '@type': type.googleapis.com/google.cloud.networkconnectivity.v1.OperationMetadata
  apiVersion: v1
  createTime: '2026-03-06T08:59:23.151624553Z'
  endTime: '2026-03-06T08:59:31.598291496Z'
  requestedCancellation: false
  target: projects/my-project/locations/us-central1/serviceConnectionPolicies/my-policy
  verb: create
name: projects/my-project/locations/us-central1/operations/operation-12344123412341-1234744f01234-c38c1234-ff31234
response:
  '@type': type.googleapis.com/google.cloud.networkconnectivity.v1.ServiceConnectionPolicy
  createTime: '2026-03-06T08:59:23.146620443Z'
  etag: T0-1234gst8kdx1234_FmdJgaNVi1234-4xNBs31234
  name: projects/my-project/locations/us-central1/serviceConnectionPolicies/my-policy
  updateTime: '2026-03-06T08:59:23.146620443Z'
*/

type ServiceConnectionPolicyOperationsConfig interface {
	ResolveServiceConnectionPolicyOperation(ctx context.Context, operationName string, opts ...OperationOptionCall) error
}

// ServiceConnectionPolicy operation resolving

func (s *store) ResolveServiceConnectionPolicyOperation(ctx context.Context, operationName string, opts ...OperationOptionCall) error {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return ctx.Err()
	}

	opName, err := gcputil.ParseNameDetail(operationName)
	if err != nil {
		return gcpmeta.NewBadRequestError("operation name is invalid: %v", err)
	}
	opBuilder, found := s.longRunningOperations.FindByName(opName)
	if !found {
		return gcpmeta.NewNotFoundError("operation %s not found", opName)
	}

	scp, found := s.serviceConnectionPolicies.FindByName(opBuilder.relatedItemName)
	if !found {
		return gcpmeta.NewNotFoundError("operation target serviceConnectionPolicy %s not found", opBuilder.relatedItemName.String())
	}
	opBuilder.WithDone(true)
	for _, op := range opts {
		if err := op(opBuilder); err != nil {
			if !errors.Is(err, common.ErrLogical) {
				err = fmt.Errorf("%w: %w", common.ErrLogical, err)
			}
			return err
		}
	}
	if err := opBuilder.WithResult(scp); err != nil {
		return err
	}

	meta, err := ReadOperationMetadata[*networkconnectivitypb.OperationMetadata](opBuilder)
	if err != nil {
		return err
	}
	if meta != nil && meta.Verb == "delete" {
		s.serviceConnectionPolicies = s.serviceConnectionPolicies.FilterNotByCallback(func(item FilterableListItem[*networkconnectivitypb.ServiceConnectionPolicy]) bool {
			return item.Name.Equal(opBuilder.relatedItemName)
		})
	}
	return nil
}

// NetworkConnectivity Operations Client Interface implementations ===================================================

func (s *store) GetNetworkConnectivityOperation(ctx context.Context, req *longrunningpb.GetOperationRequest, _ ...gax.CallOption) (*longrunningpb.Operation, error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	op, err := s.getLongRunningOperationNoLock(req.Name)
	if err != nil {
		return nil, err
	}
	return util.Clone(op)
}

func (s *store) ListNetworkConnectivityOperations(ctx context.Context, req *longrunningpb.ListOperationsRequest, _ ...gax.CallOption) gcpclient.Iterator[*longrunningpb.Operation] {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return &iteratorMocked[*longrunningpb.Operation]{
			err: ctx.Err(),
		}
	}

	return s.listLongRunningOperationsNoLock(req.Name, req.Filter)
}
