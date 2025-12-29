package client

import (
	"context"
	"fmt"
	"sync"

	"cloud.google.com/go/filestore/apiv1/filestorepb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MockFilestoreClient implements FilestoreClient for testing.
// Stores instances and operations in-memory and provides methods to control responses.
type MockFilestoreClient struct {
	mu sync.Mutex

	// Storage
	instances  map[string]*filestorepb.Instance
	operations map[string]*mockOperation

	// Error injection
	getInstanceError    error
	createInstanceError error
	updateInstanceError error
	deleteInstanceError error
	getOperationError   error

	// Operation completion control
	autoCompleteOperations bool
}

type mockOperation struct {
	name   string
	done   bool
	target string // instance name
	verb   string // create, update, delete
}

var _ FilestoreClient = &MockFilestoreClient{}

// NewMockFilestoreClient creates a new mock client for testing.
func NewMockFilestoreClient() *MockFilestoreClient {
	return &MockFilestoreClient{
		instances:              make(map[string]*filestorepb.Instance),
		operations:             make(map[string]*mockOperation),
		autoCompleteOperations: true, // By default, operations complete immediately
	}
}

func (m *MockFilestoreClient) GetInstance(ctx context.Context, projectId, location, instanceId string) (*filestorepb.Instance, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.getInstanceError != nil {
		return nil, m.getInstanceError
	}

	key := formatInstanceName(projectId, location, instanceId)
	instance, exists := m.instances[key]
	if !exists {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("instance %s not found", key))
	}

	return instance, nil
}

func (m *MockFilestoreClient) CreateInstance(ctx context.Context, projectId, location, instanceId string, instance *filestorepb.Instance) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.createInstanceError != nil {
		return "", m.createInstanceError
	}

	key := formatInstanceName(projectId, location, instanceId)

	// Check if already exists
	if _, exists := m.instances[key]; exists {
		return "", status.Error(codes.AlreadyExists, fmt.Sprintf("instance %s already exists", key))
	}

	// Store instance
	instance.Name = key
	if instance.State == filestorepb.Instance_STATE_UNSPECIFIED {
		instance.State = filestorepb.Instance_CREATING
	}
	m.instances[key] = instance

	// Create operation
	opName := fmt.Sprintf("operations/%s/create-%s", projectId, instanceId)
	op := &mockOperation{
		name:   opName,
		done:   m.autoCompleteOperations,
		target: key,
		verb:   "create",
	}
	m.operations[opName] = op

	// If auto-complete, set instance to READY
	if m.autoCompleteOperations {
		instance.State = filestorepb.Instance_READY
	}

	return opName, nil
}

func (m *MockFilestoreClient) UpdateInstance(ctx context.Context, projectId, location, instanceId string, instance *filestorepb.Instance, updateMask []string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.updateInstanceError != nil {
		return "", m.updateInstanceError
	}

	key := formatInstanceName(projectId, location, instanceId)

	// Check if exists
	existingInstance, exists := m.instances[key]
	if !exists {
		return "", status.Error(codes.NotFound, fmt.Sprintf("instance %s not found", key))
	}

	// Update only specified fields based on updateMask
	for _, field := range updateMask {
		switch field {
		case "description":
			existingInstance.Description = instance.Description
		case "labels":
			existingInstance.Labels = instance.Labels
		case "file_shares":
			existingInstance.FileShares = instance.FileShares
			// Add more fields as needed
		}
	}

	existingInstance.State = filestorepb.Instance_READY

	// Create operation
	opName := fmt.Sprintf("operations/%s/update-%s", projectId, instanceId)
	op := &mockOperation{
		name:   opName,
		done:   m.autoCompleteOperations,
		target: key,
		verb:   "update",
	}
	m.operations[opName] = op

	return opName, nil
}

func (m *MockFilestoreClient) DeleteInstance(ctx context.Context, projectId, location, instanceId string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.deleteInstanceError != nil {
		return "", m.deleteInstanceError
	}

	key := formatInstanceName(projectId, location, instanceId)

	// Check if exists
	if _, exists := m.instances[key]; !exists {
		return "", status.Error(codes.NotFound, fmt.Sprintf("instance %s not found", key))
	}

	// Delete instance
	if m.autoCompleteOperations {
		delete(m.instances, key)
	} else {
		// Mark as deleting
		m.instances[key].State = filestorepb.Instance_DELETING
	}

	// Create operation
	opName := fmt.Sprintf("operations/%s/delete-%s", projectId, instanceId)
	op := &mockOperation{
		name:   opName,
		done:   m.autoCompleteOperations,
		target: key,
		verb:   "delete",
	}
	m.operations[opName] = op

	return opName, nil
}

func (m *MockFilestoreClient) GetOperation(ctx context.Context, operationName string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.getOperationError != nil {
		return false, m.getOperationError
	}

	op, exists := m.operations[operationName]
	if !exists {
		return false, status.Error(codes.NotFound, fmt.Sprintf("operation %s not found", operationName))
	}

	return op.done, nil
}

// Test helper methods

// SetInstance directly sets an instance in the mock storage
func (m *MockFilestoreClient) SetInstance(projectId, location, instanceId string, instance *filestorepb.Instance) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := formatInstanceName(projectId, location, instanceId)
	instance.Name = key
	m.instances[key] = instance
}

// CompleteOperation marks an operation as done
func (m *MockFilestoreClient) CompleteOperation(operationName string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if op, exists := m.operations[operationName]; exists {
		op.done = true

		// Update instance state based on operation verb
		if instance, exists := m.instances[op.target]; exists {
			switch op.verb {
			case "create":
				instance.State = filestorepb.Instance_READY
			case "delete":
				delete(m.instances, op.target)
			}
		}
	}
}

// SetAutoCompleteOperations controls whether operations complete immediately
func (m *MockFilestoreClient) SetAutoCompleteOperations(autoComplete bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.autoCompleteOperations = autoComplete
}

// Error injection methods

func (m *MockFilestoreClient) SetGetInstanceError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getInstanceError = err
}

func (m *MockFilestoreClient) SetCreateInstanceError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.createInstanceError = err
}

func (m *MockFilestoreClient) SetUpdateInstanceError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.updateInstanceError = err
}

func (m *MockFilestoreClient) SetDeleteInstanceError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deleteInstanceError = err
}

func (m *MockFilestoreClient) SetGetOperationError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getOperationError = err
}

// ClearErrors resets all error injections
func (m *MockFilestoreClient) ClearErrors() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getInstanceError = nil
	m.createInstanceError = nil
	m.updateInstanceError = nil
	m.deleteInstanceError = nil
	m.getOperationError = nil
}

// Reset clears all stored instances and operations
func (m *MockFilestoreClient) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.instances = make(map[string]*filestorepb.Instance)
	m.operations = make(map[string]*mockOperation)
	m.ClearErrors()
}
