package mock2

import (
	"context"

	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	"github.com/googleapis/gax-go/v2"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
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

func (s *store) createLongRunningOperationNoLock(projectId, locationId string) *longrunningpb.Operation {
	op := &longrunningpb.Operation{
		Name: "operations/" + util.GenerateRandomString(10),
	}
	s.longRunningOperations.add(op)
	return op
}

// Client Interface implementations ===================================================

func (s *store) GetFilestoreOperation(ctx context.Context, req *longrunningpb.GetOperationRequest, _ ...gax.CallOption) (*longrunningpb.Operation, error) {
	panic("implement me")
}

func (s *store) ListFilestoreOperations(ctx context.Context, req *longrunningpb.ListOperationsRequest, _ ...gax.CallOption) gcpclient.Iterator[*longrunningpb.Operation] {
	panic("implement me")
}

func (s *store) GetRedisClusterOperation(ctx context.Context, req *longrunningpb.GetOperationRequest, _ ...gax.CallOption) (*longrunningpb.Operation, error) {
	panic("implement me")
}

func (s *store) ListRedisClusterOperations(ctx context.Context, req *longrunningpb.ListOperationsRequest, _ ...gax.CallOption) gcpclient.Iterator[*longrunningpb.Operation] {
	panic("implement me")
}

func (s *store) GetRedisInstanceOperation(ctx context.Context, req *longrunningpb.GetOperationRequest, _ ...gax.CallOption) (*longrunningpb.Operation, error) {
	panic("implement me")
}

func (s *store) ListRedisInstanceOperations(ctx context.Context, req *longrunningpb.ListOperationsRequest, _ ...gax.CallOption) gcpclient.Iterator[*longrunningpb.Operation] {
	panic("implement me")
}
