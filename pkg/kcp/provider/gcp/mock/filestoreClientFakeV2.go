package mock

import (
	"context"
	"fmt"
	"sync"

	"cloud.google.com/go/filestore/apiv1/filestorepb"
	gcpnfsinstancev2client "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsinstance/v2/client"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// FilestoreClientFakeUtils provides test utilities for manipulating v2 NfsInstance mock state
type FilestoreClientFakeUtils interface {
	// GetInstanceByName retrieves an instance directly by its full name (for testing)
	GetInstanceByName(name string) *filestorepb.Instance

	// SetInstanceReady sets the instance state to READY for testing
	SetInstanceReady(name string)
}

// filestoreClientFakeV2 implements FilestoreClient for testing.
// Stores instances and operations in-memory and provides methods to control responses.
type filestoreClientFakeV2 struct {
	mutex sync.Mutex

	// Storage
	instances  map[string]*filestorepb.Instance
	operations map[string]*mockOperationV2

	// Error injection
	getInstanceError    error
	createInstanceError error
	updateInstanceError error
	deleteInstanceError error
	getOperationError   error

	// Operation completion control
	autoCompleteOperations bool
}

type mockOperationV2 struct {
	name   string
	done   bool
	target string // instance name
	verb   string // create, update, delete
}

var _ gcpnfsinstancev2client.FilestoreClient = &filestoreClientFakeV2{}

// newFilestoreClientFakeV2 creates a new fake client for testing.
func newFilestoreClientFakeV2() *filestoreClientFakeV2 {
	return &filestoreClientFakeV2{
		instances:              make(map[string]*filestorepb.Instance),
		operations:             make(map[string]*mockOperationV2),
		autoCompleteOperations: true, // By default, operations complete immediately
	}
}

func (m *filestoreClientFakeV2) GetInstance(ctx context.Context, projectId, location, instanceId string) (*filestorepb.Instance, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.getInstanceError != nil {
		return nil, m.getInstanceError
	}

	key := formatInstanceNameV2(projectId, location, instanceId)
	instance, exists := m.instances[key]
	if !exists {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("instance %s not found", key))
	}

	return instance, nil
}

func (m *filestoreClientFakeV2) CreateInstance(ctx context.Context, projectId, location, instanceId string, instance *filestorepb.Instance) (string, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.createInstanceError != nil {
		return "", m.createInstanceError
	}

	key := formatInstanceNameV2(projectId, location, instanceId)

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
	op := &mockOperationV2{
		name:   opName,
		done:   m.autoCompleteOperations,
		target: key,
		verb:   "create",
	}
	m.operations[opName] = op

	// If auto-complete, set instance to READY
	if m.autoCompleteOperations {
		instance.State = filestorepb.Instance_READY
		// Populate Networks with IP addresses if not already set
		if len(instance.Networks) == 0 {
			instance.Networks = []*filestorepb.NetworkConfig{
				{
					IpAddresses: []string{"10.0.0.2"},
				},
			}
		} else if len(instance.Networks[0].IpAddresses) == 0 {
			instance.Networks[0].IpAddresses = []string{"10.0.0.2"}
		}
	}

	return opName, nil
}

func (m *filestoreClientFakeV2) UpdateInstance(ctx context.Context, projectId, location, instanceId string, instance *filestorepb.Instance, updateMask []string) (string, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.updateInstanceError != nil {
		return "", m.updateInstanceError
	}

	key := formatInstanceNameV2(projectId, location, instanceId)

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
	op := &mockOperationV2{
		name:   opName,
		done:   m.autoCompleteOperations,
		target: key,
		verb:   "update",
	}
	m.operations[opName] = op

	return opName, nil
}

func (m *filestoreClientFakeV2) DeleteInstance(ctx context.Context, projectId, location, instanceId string) (string, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.deleteInstanceError != nil {
		return "", m.deleteInstanceError
	}

	key := formatInstanceNameV2(projectId, location, instanceId)

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
	op := &mockOperationV2{
		name:   opName,
		done:   m.autoCompleteOperations,
		target: key,
		verb:   "delete",
	}
	m.operations[opName] = op

	return opName, nil
}

func (m *filestoreClientFakeV2) GetOperation(ctx context.Context, operationName string) (bool, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.getOperationError != nil {
		return false, m.getOperationError
	}

	op, exists := m.operations[operationName]
	if !exists {
		return false, status.Error(codes.NotFound, fmt.Sprintf("operation %s not found", operationName))
	}

	return op.done, nil
}

// Test helper methods (Utils interface implementation)

// GetInstanceByName retrieves an instance directly by its full name (for testing)
func (m *filestoreClientFakeV2) GetInstanceByName(name string) *filestorepb.Instance {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.instances[name]
}

// SetInstanceReady sets the instance state to READY for testing
func (m *filestoreClientFakeV2) SetInstanceReady(name string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if instance, ok := m.instances[name]; ok {
		instance.State = filestorepb.Instance_READY
	}
}

// formatInstanceNameV2 constructs the full GCP instance name
func formatInstanceNameV2(projectId, location, instanceId string) string {
	return fmt.Sprintf("projects/%s/locations/%s/instances/%s", projectId, location, instanceId)
}
