package mock2

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/google/uuid"
	"github.com/googleapis/gax-go/v2"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
	gcputil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
)

/*

endTime: '2026-03-16T01:40:42.904-07:00'
id: '65472920581234084321'
insertTime: '2026-03-16T01:40:29.135-07:00'
kind: compute#operation
name: operation-17736504212343-64d1234c41234-69812342-374b1234
operationType: insert
progress: 100
selfLink: https://www.googleapis.com/compute/v1/projects/my-project/global/operations/operation-17736504212343-64d1234c41234-69812342-374b1234
startTime: '2026-03-16T01:40:29.138-07:00'
status: DONE
targetId: '90102244315434612340'
targetLink: https://www.googleapis.com/compute/v1/projects/my-project/global/networks/my-network
user: someone@sap.com

---

endTime: '2026-03-02T02:49:16.801-08:00'
id: '4469345956290680260'
insertTime: '2026-03-02T02:49:15.504-08:00'
kind: compute#operation
name: operation-1123448555089-64c085612343c-00d51234-e1cb1234
operationType: addPeering
progress: 100
selfLink: https://www.googleapis.com/compute/v1/projects/my-project/global/operations/operation-1123448555089-64c085612343c-00d51234-e1cb1234
startTime: '2026-03-02T02:49:15.507-08:00'
status: DONE
targetId: '6922114947784171892'
targetLink: https://www.googleapis.com/compute/v1/projects/my-project/global/networks/my-network
user: someone@my-project.iam.gserviceaccount.com

*/

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

func (s *store) ResolveComputeOperation(nd gcputil.NameDetail, httpErrorStatusCode *int32, httpErrorMessage *string, err *computepb.Error) error {
	s.m.Lock()
	defer s.m.Unlock()

	op, found := s.computeOperations.FindByName(nd)
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
	op.HttpErrorStatusCode = httpErrorStatusCode
	op.HttpErrorMessage = httpErrorMessage
	return nil
}

// ==============================================================================

func (s *store) GetComputeGlobalOperation(ctx context.Context, req *computepb.GetGlobalOperationRequest, _ ...gax.CallOption) (*computepb.Operation, error) {
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

	op, err := s.getComputeOperationNoLock(gcputil.NewGlobalOperationName(req.Project, req.Operation))
	if err != nil {
		return nil, err
	}
	return util.Clone(op)
}

func (s *store) ListComputeGlobalOperations(ctx context.Context, req *computepb.ListGlobalOperationsRequest, _ ...gax.CallOption) gcpclient.Iterator[*computepb.Operation] {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return &iteratorMocked[*computepb.Operation]{err: ctx.Err()}
	}

	list := s.computeOperations.FilterByCallback(func(l FilterableListItem[*computepb.Operation]) bool {
		return l.Obj.Region == nil && l.Name.ProjectId() == req.Project
	})
	var err error
	list, err = list.FilterByExpression(req.Filter)
	if err != nil {
		return &iteratorMocked[*computepb.Operation]{err: err}
	}

	return list.ToIterator()
}

func (s *store) GetComputeRegionalOperation(ctx context.Context, req *computepb.GetRegionOperationRequest, _ ...gax.CallOption) (*computepb.Operation, error) {
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

	op, err := s.getComputeOperationNoLock(gcputil.NewLocationalOperationName(req.Project, req.Region, req.Operation))
	if err != nil {
		return nil, err
	}
	return util.Clone(op)
}

func (s *store) ListComputeRegionalOperations(ctx context.Context, req *computepb.ListRegionOperationsRequest, _ ...gax.CallOption) gcpclient.Iterator[*computepb.Operation] {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return &iteratorMocked[*computepb.Operation]{err: ctx.Err()}
	}

	list := s.computeOperations.FilterByCallback(func(l FilterableListItem[*computepb.Operation]) bool {
		return l.Obj.Region != nil && l.Name.ProjectId() == req.Project && l.Name.LocationRegionId() == req.Region
	})
	var err error
	list, err = list.FilterByExpression(req.Filter)
	if err != nil {
		return &iteratorMocked[*computepb.Operation]{err: err}
	}

	return list.ToIterator()
}

// ==============================================================================

func (s *store) getComputeOperationNoLock(nd gcputil.NameDetail) (*computepb.Operation, error) {
	op, found := s.computeOperations.FindByName(nd)
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
		nd = gcputil.NewLocationalOperationName(projectId, region, "operation-"+uuid.NewString())
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

	s.computeOperations.Add(op, nd)

	return op
}
