package mock2

import (
	"context"
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/google/uuid"
	"github.com/googleapis/gax-go/v2"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
	gcputil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"github.com/mitchellh/copystructure"
	"k8s.io/utils/ptr"
)

func reactComputeOperationDone(op *computepb.Operation) bool {
	op.Status = ptr.To(computepb.Operation_DONE)
	op.Progress = ptr.To(int32(100))
	op.EndTime = ptr.To(time.Now().Format(time.RFC3339))
	return true
}

func ReactComputeOperationDoneWithError(namePart string, err *computepb.Error) ReactorFunc[*computepb.Operation] {
	return func(op *computepb.Operation) bool {
		if namePart == "" || strings.Contains(ptr.Deref(op.SelfLink, ""), namePart) {
			op.Status = ptr.To(computepb.Operation_DONE)
			op.Progress = ptr.To(int32(100))
			op.EndTime = ptr.To(time.Now().Format(time.RFC3339))
			op.Error = err
			return true
		}
		return false
	}
}

func NewSimpleComputeOperationError(code, message string) *computepb.Error {
	return &computepb.Error{
		Errors: []*computepb.Errors{
			{
				Code:    ptr.To(code),
				Message: ptr.To(message),
			},
		},
	}
}

// AddReactorComputeOperation adds a new reactor for compute operations. By default, the only reactor added will resolve
// each operation as done w/out error. Use this func to add custom specific reactors, e.g. to resolve only specific
// operations or to resolve them with an error. The operation has its target so you can use that to only resolve specific
// operations. If your reactor handles the operation it should return true and that will stop execution of other reactors,
// including the default one that will resolve it to done w/out error.
func (s *store) AddReactorComputeOperation(r ReactorFunc[*computepb.Operation]) {
	s.m.Lock()
	defer s.m.Unlock()

	s.reactorComputeOperations.AddReactor(r)
}

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

func (s *store) ResolveComputeOperation(nd gcputil.NameDetail, err *computepb.Error) error {
	s.m.Lock()
	defer s.m.Unlock()

	op, found := s.computeOperations.findByName(nd)
	if !found {
		return gcpmeta.NewNotFoundError("operation %s not found", nd.String())
	}
	if ptr.Deref(op.Status, computepb.Operation_UNDEFINED_STATUS) == computepb.Operation_DONE {
		return gcpmeta.NewBadRequestError("operation %s is already done", nd.String())
	}
	op.Status = ptr.To(computepb.Operation_DONE)
	op.EndTime = ptr.To(time.Now().Format(time.RFC3339))
	op.Progress = ptr.To(int32(100))
	op.Error = err
	return nil
}

// ==============================================================================

func (s *store) GetComputeGlobalOperation(ctx context.Context, req *computepb.GetGlobalOperationRequest, opts ...gax.CallOption) (*computepb.Operation, error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	if req.Project == "" {
		return nil, gcpmeta.NewBadRequestError("project is required")
	}
	if req.Operation == "" {
		return nil, gcpmeta.NewBadRequestError("operation is required")
	}

	return s.getComputeOperationNoLock(gcputil.NewGlobalOperationName(req.Project, req.Operation))
}

func (s *store) ListComputeGlobalOperations(ctx context.Context, req *computepb.ListGlobalOperationsRequest, opts ...gax.CallOption) gcpclient.Iterator[*computepb.Operation] {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return &iteratorMocked[*computepb.Operation]{err: ctx.Err()}
	}

	list := s.computeOperations.filterByCallback(func(l listItem[*computepb.Operation]) bool {
		return l.obj.Region == nil && l.name.ProjectId() == req.Project
	})
	var err error
	list, err = s.computeOperations.filterByExpression(req.Filter)
	if err != nil {
		return &iteratorMocked[*computepb.Operation]{err: err}
	}

	return list.toIterator()
}

func (s *store) GetComputeRegionalOperation(ctx context.Context, req *computepb.GetRegionOperationRequest, opts ...gax.CallOption) (*computepb.Operation, error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	if req.Project == "" {
		return nil, gcpmeta.NewBadRequestError("project is required")
	}
	if req.Operation == "" {
		return nil, gcpmeta.NewBadRequestError("operation is required")
	}
	if req.Region == "" {
		return nil, gcpmeta.NewBadRequestError("region is required")
	}

	op, err := s.getComputeOperationNoLock(gcputil.NewRegionalOperationName(req.Project, req.Region, req.Operation))
	if err != nil {
		return nil, err
	}
	cpy, err := copystructure.Copy(op)
	if err != nil {
		return nil, fmt.Errorf("failed to copy operation: %w", err)
	}
	return cpy.(*computepb.Operation), nil
}

func (s *store) ListComputeRegionalOperations(ctx context.Context, req *computepb.ListRegionOperationsRequest, opts ...gax.CallOption) gcpclient.Iterator[*computepb.Operation] {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return &iteratorMocked[*computepb.Operation]{err: ctx.Err()}
	}

	list := s.computeOperations.filterByCallback(func(l listItem[*computepb.Operation]) bool {
		return l.obj.Region != nil && l.name.ProjectId() == req.Project && l.name.LocationRegionId() == req.Region
	})
	var err error
	list, err = list.filterByExpression(req.Filter)
	if err != nil {
		return &iteratorMocked[*computepb.Operation]{err: err}
	}

	return list.toIterator()
}

// ==============================================================================

func (s *store) getComputeOperationNoLock(nd gcputil.NameDetail) (*computepb.Operation, error) {
	op, found := s.computeOperations.findByName(nd)
	if !found {
		return nil, gcpmeta.NewNotFoundError("operation %s not found", nd.String())
	}
	return op, nil
}

func (s *store) createComputeOperationNoLock(projectId, region, operationType, targetLink string, targetId uint64) *computepb.Operation {
	var nd gcputil.NameDetail
	if region == "" {
		nd = gcputil.NewGlobalOperationName(projectId, "operation-"+uuid.NewString())
	} else {
		nd = gcputil.NewRegionalOperationName(projectId, region, "operation-"+uuid.NewString())
	}
	id := rand.Uint64()
	op := &computepb.Operation{
		Id:            ptr.To(id),
		InsertTime:    ptr.To(time.Now().Format(time.RFC3339)),
		Kind:          ptr.To("compute#operation"),
		Name:          ptr.To(nd.ResourceId()),
		OperationType: ptr.To(operationType),
		Progress:      ptr.To(int32(1)),
		SelfLink:      ptr.To(nd.PrefixWithGoogleApisComputeV1()),
		Status:        ptr.To(computepb.Operation_RUNNING),
		TargetId:      ptr.To(targetId),
		TargetLink:    ptr.To(targetLink),
		User:          ptr.To(fmt.Sprintf("user@%s.iam.gserviceaccount.com", projectId)),
	}
	if region != "" {
		op.Region = ptr.To(gcputil.NewRegionName(projectId, region).PrefixWithGoogleApisComputeV1())
	}

	s.computeOperations.add(op, nd)
	s.reactorComputeOperations.React(op)
	return op
}
