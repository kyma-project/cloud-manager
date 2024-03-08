package client

import (
	"context"
	"fmt"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/option"
	"google.golang.org/api/transport"
	"net/http"
	"sync"
)

type ClientProvider[T any] func(ctx context.Context, saJsonKeyPath string) (T, error)

func NewCachedClientProvider[T comparable](p ClientProvider[T]) ClientProvider[T] {
	var result T
	var nilT T
	var m sync.Mutex
	return func(ctx context.Context, saJsonKeyPath string) (T, error) {
		m.Lock()
		defer m.Unlock()
		var err error
		if result == nilT {
			result, err = p(ctx, saJsonKeyPath)
		}
		return result, err
	}
}

var gcpClient *http.Client
var clientMutex sync.Mutex

func newCachedGcpClient(ctx context.Context, saJsonKeyPath string) (*http.Client, error) {
	clientMutex.Lock()
	defer clientMutex.Unlock()
	if gcpClient == nil {
		client, _, err := transport.NewHTTPClient(ctx, option.WithCredentialsFile(saJsonKeyPath), option.WithScopes("https://www.googleapis.com/auth/compute", "https://www.googleapis.com/auth/cloud-platform"))
		if err != nil {
			return nil, fmt.Errorf("error obtaining GCP connection: [%w]", err)
		}
		gcpClient = client
	}
	return gcpClient, nil
}

func GetCachedGcpClient(ctx context.Context, saJsonKeyPath string) (*http.Client, error) {
	if gcpClient == nil {
		return newCachedGcpClient(ctx, saJsonKeyPath)
	}
	return gcpClient, nil
}

var projectNumbers map[string]int64 = make(map[string]int64)
var projectNumbersMutex sync.Mutex

// GetCachedProjectNumber get project number from cloud resources manager for a given project id
func GetCachedProjectNumber(ctx context.Context, projectId string, crmService *cloudresourcemanager.Service) (int64, error) {
	projectNumbersMutex.Lock()
	defer projectNumbersMutex.Unlock()
	projectNumber, ok := projectNumbers[projectId]
	if !ok {
		project, err := crmService.Projects.Get(projectId).Do()
		if err != nil {
			return 0, err
		}
		projectNumbers[projectId] = project.ProjectNumber
		projectNumber = project.ProjectNumber
	}
	return projectNumber, nil
}
