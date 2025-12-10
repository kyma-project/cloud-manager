package client

import (
	"context"

	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/compute/v1"
)

// LegacyComputeClient provides the OLD interface for backward compatibility with v2 code.
// This adapter converts between NEW pattern types (computepb.*) and OLD pattern types (compute.v1.*).
// TEMPORARY: This adapter will be removed in Phase 4 when v2/ directory is eliminated.
type LegacyComputeClient interface {
	ListGlobalAddresses(ctx context.Context, projectId, vpc string) (*compute.AddressList, error)
	CreatePscIpRange(ctx context.Context, projectId, vpcName, name, description, address string, prefixLength int64) (*compute.Operation, error)
	DeleteIpRange(ctx context.Context, projectId, name string) (*compute.Operation, error)
	GetIpRange(ctx context.Context, projectId, name string) (*compute.Address, error)
	GetGlobalOperation(ctx context.Context, projectId, operationName string) (*compute.Operation, error)
}

// NewLegacyComputeClient wraps a NEW pattern ComputeClient to provide OLD interface.
func NewLegacyComputeClient(newClient ComputeClient) LegacyComputeClient {
	return &legacyComputeClientAdapter{newClient: newClient}
}

type legacyComputeClientAdapter struct {
	newClient ComputeClient
}

func (a *legacyComputeClientAdapter) GetIpRange(ctx context.Context, projectId, name string) (*compute.Address, error) {
	addr, err := a.newClient.GetIpRange(ctx, projectId, name)
	if err != nil {
		return nil, err
	}
	return convertAddressToLegacy(addr), nil
}

func (a *legacyComputeClientAdapter) DeleteIpRange(ctx context.Context, projectId, name string) (*compute.Operation, error) {
	opName, err := a.newClient.DeleteIpRange(ctx, projectId, name)
	if err != nil {
		return nil, err
	}

	// Get the full operation to return
	op, err := a.newClient.GetGlobalOperation(ctx, projectId, opName)
	if err != nil {
		return nil, err
	}

	return convertOperationToLegacy(op), nil
}

func (a *legacyComputeClientAdapter) CreatePscIpRange(ctx context.Context, projectId, vpcName, name, description, address string, prefixLength int64) (*compute.Operation, error) {
	opName, err := a.newClient.CreatePscIpRange(ctx, projectId, vpcName, name, description, address, prefixLength)
	if err != nil {
		return nil, err
	}

	// Get the full operation to return
	op, err := a.newClient.GetGlobalOperation(ctx, projectId, opName)
	if err != nil {
		return nil, err
	}

	return convertOperationToLegacy(op), nil
}

func (a *legacyComputeClientAdapter) ListGlobalAddresses(ctx context.Context, projectId, vpc string) (*compute.AddressList, error) {
	addresses, err := a.newClient.ListGlobalAddresses(ctx, projectId, vpc)
	if err != nil {
		return nil, err
	}

	legacyAddresses := make([]*compute.Address, len(addresses))
	for i, addr := range addresses {
		legacyAddresses[i] = convertAddressToLegacy(addr)
	}

	return &compute.AddressList{
		Items: legacyAddresses,
	}, nil
}

func (a *legacyComputeClientAdapter) GetGlobalOperation(ctx context.Context, projectId, operationName string) (*compute.Operation, error) {
	op, err := a.newClient.GetGlobalOperation(ctx, projectId, operationName)
	if err != nil {
		return nil, err
	}
	return convertOperationToLegacy(op), nil
}

// Helper functions to convert from NEW types to OLD types

func ptrString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func ptrInt32(i *int32) int32 {
	if i == nil {
		return 0
	}
	return *i
}

func ptrUint64(i *uint64) uint64 {
	if i == nil {
		return 0
	}
	return *i
}

func convertAddressToLegacy(addr *computepb.Address) *compute.Address {
	if addr == nil {
		return nil
	}

	return &compute.Address{
		Name:              ptrString(addr.Name),
		Description:       ptrString(addr.Description),
		Address:           ptrString(addr.Address),
		PrefixLength:      int64(ptrInt32(addr.PrefixLength)),
		Network:           ptrString(addr.Network),
		AddressType:       ptrString(addr.AddressType),
		Purpose:           ptrString(addr.Purpose),
		Status:            ptrString(addr.Status),
		Region:            ptrString(addr.Region),
		SelfLink:          ptrString(addr.SelfLink),
		Id:                ptrUint64(addr.Id),
		CreationTimestamp: ptrString(addr.CreationTimestamp),
	}
}

func convertOperationToLegacy(op *computepb.Operation) *compute.Operation {
	if op == nil {
		return nil
	}

	status := "PENDING"
	if op.Status != nil {
		status = op.Status.String()
	}

	var httpErrorMessage string
	var httpErrorStatusCode int64
	if op.HttpErrorMessage != nil {
		httpErrorMessage = *op.HttpErrorMessage
	}
	if op.HttpErrorStatusCode != nil {
		httpErrorStatusCode = int64(*op.HttpErrorStatusCode)
	}

	legacyOp := &compute.Operation{
		Name:                ptrString(op.Name),
		Status:              status,
		StatusMessage:       ptrString(op.StatusMessage),
		TargetLink:          ptrString(op.TargetLink),
		Progress:            int64(ptrInt32(op.Progress)),
		Region:              ptrString(op.Region),
		SelfLink:            ptrString(op.SelfLink),
		HttpErrorMessage:    httpErrorMessage,
		HttpErrorStatusCode: httpErrorStatusCode,
		OperationType:       ptrString(op.OperationType),
	}

	// Convert error if present
	if op.Error != nil && len(op.Error.Errors) > 0 {
		legacyOp.Error = &compute.OperationError{
			Errors: make([]*compute.OperationErrorErrors, len(op.Error.Errors)),
		}
		for i, err := range op.Error.Errors {
			legacyOp.Error.Errors[i] = &compute.OperationErrorErrors{
				Code:    ptrString(err.Code),
				Message: ptrString(err.Message),
			}
		}
	}

	return legacyOp
}
