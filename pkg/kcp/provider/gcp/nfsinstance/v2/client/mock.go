package client

import (
	"context"

	"google.golang.org/api/file/v1"
)

// mockFilestoreClient provides a mock implementation of FilestoreClient for testing.
type mockFilestoreClient struct {
	getInstance    func(ctx context.Context, projectId, location, instanceId string) (*file.Instance, error)
	createInstance func(ctx context.Context, projectId, location, instanceId string, instance *file.Instance) (*file.Operation, error)
	updateInstance func(ctx context.Context, projectId, location, instanceId, updateMask string, instance *file.Instance) (*file.Operation, error)
	deleteInstance func(ctx context.Context, projectId, location, instanceId string) (*file.Operation, error)
	getOperation   func(ctx context.Context, projectId, operationName string) (*file.Operation, error)
}

// NewMockFilestoreClient creates a new mock client.
func NewMockFilestoreClient() *mockFilestoreClient {
	return &mockFilestoreClient{}
}

// SetGetInstance sets the mock behavior for GetInstance.
func (m *mockFilestoreClient) SetGetInstance(fn func(ctx context.Context, projectId, location, instanceId string) (*file.Instance, error)) {
	m.getInstance = fn
}

// SetCreateInstance sets the mock behavior for CreateInstance.
func (m *mockFilestoreClient) SetCreateInstance(fn func(ctx context.Context, projectId, location, instanceId string, instance *file.Instance) (*file.Operation, error)) {
	m.createInstance = fn
}

// SetUpdateInstance sets the mock behavior for UpdateInstance.
func (m *mockFilestoreClient) SetUpdateInstance(fn func(ctx context.Context, projectId, location, instanceId, updateMask string, instance *file.Instance) (*file.Operation, error)) {
	m.updateInstance = fn
}

// SetDeleteInstance sets the mock behavior for DeleteInstance.
func (m *mockFilestoreClient) SetDeleteInstance(fn func(ctx context.Context, projectId, location, instanceId string) (*file.Operation, error)) {
	m.deleteInstance = fn
}

// SetGetOperation sets the mock behavior for GetOperation.
func (m *mockFilestoreClient) SetGetOperation(fn func(ctx context.Context, projectId, operationName string) (*file.Operation, error)) {
	m.getOperation = fn
}

// GetInstance implements FilestoreClient.
func (m *mockFilestoreClient) GetInstance(ctx context.Context, projectId, location, instanceId string) (*file.Instance, error) {
	if m.getInstance != nil {
		return m.getInstance(ctx, projectId, location, instanceId)
	}
	return nil, nil
}

// CreateInstance implements FilestoreClient.
func (m *mockFilestoreClient) CreateInstance(ctx context.Context, projectId, location, instanceId string, instance *file.Instance) (*file.Operation, error) {
	if m.createInstance != nil {
		return m.createInstance(ctx, projectId, location, instanceId, instance)
	}
	return &file.Operation{Done: true}, nil
}

// UpdateInstance implements FilestoreClient.
func (m *mockFilestoreClient) UpdateInstance(ctx context.Context, projectId, location, instanceId, updateMask string, instance *file.Instance) (*file.Operation, error) {
	if m.updateInstance != nil {
		return m.updateInstance(ctx, projectId, location, instanceId, updateMask, instance)
	}
	return &file.Operation{Done: true}, nil
}

// DeleteInstance implements FilestoreClient.
func (m *mockFilestoreClient) DeleteInstance(ctx context.Context, projectId, location, instanceId string) (*file.Operation, error) {
	if m.deleteInstance != nil {
		return m.deleteInstance(ctx, projectId, location, instanceId)
	}
	return &file.Operation{Done: true}, nil
}

// GetOperation implements FilestoreClient.
func (m *mockFilestoreClient) GetOperation(ctx context.Context, projectId, operationName string) (*file.Operation, error) {
	if m.getOperation != nil {
		return m.getOperation(ctx, projectId, operationName)
	}
	return &file.Operation{Done: true}, nil
}
