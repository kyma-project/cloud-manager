package mock2

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	"cloud.google.com/go/redis/apiv1/redispb"
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
  '@type': type.googleapis.com/google.cloud.redis.v1.OperationMetadata
  apiVersion: v1
  cancelRequested: false
  createTime: '2026-03-02T05:24:19.168269921Z'
  endTime: '2026-03-02T05:24:55.850881656Z'
  target: projects/my-project/locations/europe-west1/instances/my-instance
  verb: update
name: projects/my-project/locations/europe-west1/operations/operation-1772429059113-64c03cc75420f-1e6dfa66-a66a511c
response:
  '@type': type.googleapis.com/google.cloud.redis.v1.Instance
  authorizedNetwork: projects/my-project/global/networks/my-network
  availableMaintenanceVersions:
  - '20251110_01_00'
  connectMode: PRIVATE_SERVICE_ACCESS
  createTime: '2025-03-31T17:44:56.236382111Z'
  currentLocationId: europe-west1-c
  host: 172.16.0.19
  locationId: europe-west1-c
  maintenanceVersion: '20251007_00_00'
  memorySizeGb: 2
  name: projects/my-project/locations/europe-west1/instances/my-instance
  nodes:
  - id: node-0
    zone: europe-west1-c
  persistenceConfig:
    persistenceMode: DISABLED
  persistenceIamIdentity: serviceAccount:service-123412341234@cloud-redis.iam.gserviceaccount.com
  port: 6378
  readReplicasMode: READ_REPLICAS_DISABLED
  redisConfigs:
    maxmemory-policy: noeviction
  redisVersion: REDIS_7_0
  reservedIpRange: 172.16.0.0/29
  satisfiesPzi: true
  serverCaCerts:
  - cert: |-
      -----BEGIN CERTIFICATE-----
      ...
      -----END CERTIFICATE-----
    createTime: '2025-03-31T17:45:04.218873Z'
    expireTime: '2035-03-29T17:45:04.073Z'
    serialNumber: '0'
    sha1Fingerprint: 123a1a23c3f21c4b21c21f0a0a0c123f21b0a21
  state: READY
  tier: BASIC
  transitEncryptionMode: SERVER_AUTHENTICATION
*/

type RedisInstanceOperationsConfig interface {
	ResolveRedisInstanceOperation(ctx context.Context, operationName string, opts ...RedisInstanceOperationOptionCall) error
}

// RedisInstance operation resolving

type RedisInstanceOperationOptionCall func(ri *redispb.Instance, opBuilder *OperationLongRunningBuilder) error

func WithRedisInstanceOperationSimpleError(code int32, message string) RedisInstanceOperationOptionCall {
	return func(ri *redispb.Instance, opBuilder *OperationLongRunningBuilder) error {
		ri.State = redispb.Instance_STATE_UNSPECIFIED // TODO ??? there's no error state
		opBuilder.WithSimpleError(code, message)
		return nil
	}
}

func WithRedisInstanceOperationError(err *longrunningpb.Operation_Error) RedisInstanceOperationOptionCall {
	return func(ri *redispb.Instance, opBuilder *OperationLongRunningBuilder) error {
		ri.State = redispb.Instance_STATE_UNSPECIFIED // TODO ??? there's no error state
		opBuilder.WithOperationError(err)
		return nil
	}
}

func (s *store) ResolveRedisInstanceOperation(ctx context.Context, operationName string, opts ...RedisInstanceOperationOptionCall) error {
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

	ri, found := s.redisInstances.FindByName(opBuilder.relatedItemName)
	if !found {
		return gcpmeta.NewNotFoundError("operation target redisInstance %s not found", opBuilder.relatedItemName.String())
	}
	ri.State = redispb.Instance_READY
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

	meta, err := ReadOperationMetadata[*redispb.OperationMetadata](opBuilder)
	if err != nil {
		return err
	}
	if meta != nil && meta.Verb == "delete" {
		s.redisInstances = s.redisInstances.FilterNotByCallback(func(item FilterableListItem[*redispb.Instance]) bool {
			return item.Name.Equal(opBuilder.relatedItemName)
		})
	}
	return nil
}

// Client Interface implementations ===================================================

func (s *store) GetRedisInstanceOperation(ctx context.Context, req *longrunningpb.GetOperationRequest, _ ...gax.CallOption) (*longrunningpb.Operation, error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	return s.getLongRunningOperationNoLock(req.Name)
}

func (s *store) ListRedisInstanceOperations(ctx context.Context, req *longrunningpb.ListOperationsRequest, _ ...gax.CallOption) gcpclient.Iterator[*longrunningpb.Operation] {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return &iteratorMocked[*longrunningpb.Operation]{
			err: ctx.Err(),
		}
	}

	return s.listLongRunningOperationsNoLock(req.Name, req.Filter)
}
