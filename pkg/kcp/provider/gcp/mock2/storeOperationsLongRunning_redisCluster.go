package mock2

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	"cloud.google.com/go/redis/cluster/apiv1/clusterpb"
	"github.com/googleapis/gax-go/v2"
	"github.com/kyma-project/cloud-manager/pkg/common"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
	gcputil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

/*
done: false
metadata:
  '@type': type.googleapis.com/google.cloud.redis.cluster.v1alpha1.OperationMetadata
  apiVersion: v1alpha1
  createTime: '2026-03-09T12:18:43.062638389Z'
  requestedCancellation: false
  target: projects/sap-sc-learn/locations/us-central1/clusters/test-tmp
  verb: create
name: projects/sap-sc-learn/locations/us-central1/operations/operation-1773058722854-64c966757756d-94a8b4ad-1276e7f3
 */

type RedisClusterOperationsConfig interface {
	ResolveRedisClusterOperation(ctx context.Context, operationName string, opts ...RedisClusterOperationOptionCall) error
}

// RedisInstance operation resolving

type RedisClusterOperationOptionCall func(rc *clusterpb.Cluster, opBuilder *OperationLongRunningBuilder) error

func WithRedisClusterOperationSimpleError(code int32, message string) RedisClusterOperationOptionCall {
	return func(rc *clusterpb.Cluster, opBuilder *OperationLongRunningBuilder) error {
		rc.State = clusterpb.Cluster_STATE_UNSPECIFIED // TODO ??? there's no error state
		opBuilder.WithSimpleError(code, message)
		return nil
	}
}

func WithRedisClusterOperationError(err *longrunningpb.Operation_Error) RedisClusterOperationOptionCall {
	return func(rc *clusterpb.Cluster, opBuilder *OperationLongRunningBuilder) error {
		rc.State = clusterpb.Cluster_STATE_UNSPECIFIED // TODO ??? there's no error state
		opBuilder.WithOperationError(err)
		return nil
	}
}

func (s *store) ResolveRedisClusterOperation(ctx context.Context, operationName string, opts ...RedisClusterOperationOptionCall) error {
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

	ri, found := s.redisClusters.FindByName(opBuilder.relatedItemName)
	if !found {
		return gcpmeta.NewNotFoundError("operation target redisCluster %s not found", opBuilder.relatedItemName.String())
	}
	ri.State = clusterpb.Cluster_ACTIVE
	opBuilder.WithDone(true)
	for _, op := range opts {
		if err := op(ri, opBuilder); err != nil {
			if !errors.Is(err, common.ErrLogical) {
				err = fmt.Errorf("%w: %w", common.ErrLogical, err)
			}
			return err
		}
	}
	if err := opBuilder.WithResult(ri); err != nil {
		return err
	}

	meta, err := ReadOperationMetadata[*clusterpb.OperationMetadata](opBuilder)
	if err != nil {
		return err
	}
	if meta != nil && meta.Verb == "delete" {
		s.redisClusters = s.redisClusters.FilterNotByCallback(func(item FilterableListItem[*clusterpb.Cluster]) bool {
			return item.Name.Equal(opBuilder.relatedItemName)
		})
	}
	return nil
}

// Client Interface implementations ===================================================

func (s *store) GetRedisClusterOperation(ctx context.Context, req *longrunningpb.GetOperationRequest, _ ...gax.CallOption) (*longrunningpb.Operation, error) {
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
