package mock2

import (
	"fmt"

	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	"github.com/google/uuid"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
	gcputil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/util"
)

func (s *store) newLongRunningOperationName(projectId string) gcputil.NameDetail {
	return gcputil.NewOperationName(projectId, fmt.Sprintf("operation-%s", uuid.NewString()))
}

// base generic get&list methods =====================================================

func (s *store) getLongRunningOperationNoLock(name string) (*longrunningpb.Operation, error) {
	nd, err := gcputil.ParseNameDetail(name)
	if err != nil {
		return nil, gcpmeta.NewBadRequestError("invalid operation name: %v", err)
	}
	op, found := s.longRunningOperations.FindByName(nd)
	if !found {
		return nil, gcpmeta.NewNotFoundError("operation %s not found", name)
	}
	return op.GetOperationPB(), nil
}

func (s *store) longRunningOperationsProjected() (*FilterableList[*longrunningpb.Operation], error) {
	result, err := MapFilterableList[*OperationLongRunningBuilder, *longrunningpb.Operation](
		s.longRunningOperations,
		func(a *OperationLongRunningBuilder) *longrunningpb.Operation {
			return a.GetOperationPB()
		},
		nil,
	)
	return result, err
}

func (s *store) listLongRunningOperationsNoLock(parentName, filter string) gcpclient.Iterator[*longrunningpb.Operation] {
	result, err := s.longRunningOperationsProjected()
	if err != nil {
		return &iteratorMocked[*longrunningpb.Operation]{
			err: gcpmeta.NewBadRequestError("%v: failed to map builders to operations", err),
		}
	}

	if parentName != "" {
		parentNd, err := gcputil.ParseNameDetail(parentName)
		if err != nil {
			return &iteratorMocked[*longrunningpb.Operation]{
				err: gcpmeta.NewBadRequestError("invalid parent name: %v", err),
			}
		}
		result = result.FilterByParent(parentNd)
	}

	if filter != "" {
		var err error
		result, err = result.FilterByExpression(&filter)
		if err != nil {
			return &iteratorMocked[*longrunningpb.Operation]{
				err: gcpmeta.NewBadRequestError("invalid filter: %v", err),
			}
		}
	}

	return result.ToIterator()
}
