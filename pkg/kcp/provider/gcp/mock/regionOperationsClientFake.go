package mock

import (
	"context"
	"sync"

	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/subnet/client"
	"google.golang.org/api/googleapi"
)

type RegionalOperationsClientFakeUtils interface {
	AddRegionOperation(name string) string
	GetRegionOperationById(operationId string) *computepb.Operation
	SetRegionOperationDone(operationId string)
	SetRegionOperationError(operationId string)
}

type regionalOperationsClientFake struct {
	mutex      sync.Mutex
	operations map[string]*computepb.Operation
}

func (c *regionalOperationsClientFake) AddRegionOperation(name string) string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	key := name
	c.operations[key] = &computepb.Operation{
		Status: computepb.Operation_PENDING.Enum(),
		Name:   &key,
	}
	return name
}

func (c *regionalOperationsClientFake) GetRegionOperationById(operationId string) *computepb.Operation {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	op := c.operations[operationId]
	return op
}

func (c *regionalOperationsClientFake) SetRegionOperationDone(operationId string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	op := c.operations[operationId]

	op.Status = computepb.Operation_DONE.Enum()
}

func (c *regionalOperationsClientFake) SetRegionOperationError(operationId string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	op := c.operations[operationId]

	op.Status = computepb.Operation_DONE.Enum()
	op.Error = &computepb.Error{Errors: []*computepb.Errors{}}
}

func (c *regionalOperationsClientFake) GetRegionOperation(ctx context.Context, request client.GetRegionOperationRequest) (*computepb.Operation, error) {
	select {
	case <-ctx.Done():
		return nil, context.Canceled
	default:
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	key := request.Name

	op, ok := c.operations[key]
	if !ok {
		return nil, &googleapi.Error{
			Code:    404,
			Message: "Not Found",
		}
	}

	return op, nil
}
