package client

import (
	"context"
	"fmt"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/option"
	"google.golang.org/api/transport"
	"net/http"
	"os"
	"os/signal"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sync"
	"syscall"
	"time"
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
		client, err := newHttpClient(ctx, saJsonKeyPath)
		if err != nil {
			return nil, err
		}
		gcpClient = client
		go renewCachedHttpClientPeriodically(context.Background(), saJsonKeyPath, os.Getenv("GCP_CLIENT_RENEW_DURATION"))
	}
	return gcpClient, nil
}

func GetCachedGcpClient(ctx context.Context, saJsonKeyPath string) (*http.Client, error) {
	if gcpClient == nil {
		return newCachedGcpClient(ctx, saJsonKeyPath)
	}
	return gcpClient, nil
}

var projectNumbers = make(map[string]int64)
var projectNumbersMutex sync.Mutex

// GetCachedProjectNumber get project number from cloud resources manager for a given project id
func GetCachedProjectNumber(projectId string, crmService *cloudresourcemanager.Service) (int64, error) {
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

func renewCachedHttpClientPeriodically(ctx context.Context, saJsonKeyPath, duration string) {
	logger := log.FromContext(ctx)
	if duration == "" {
		logger.Info("GCP_CLIENT_RENEW_DURATION not set, defaulting to 5m")
		duration = "5m"
	}
	period, err := time.ParseDuration(duration)
	if err != nil {
		logger.Error(err, "error parsing GCP_CLIENT_RENEW_DURATION, defaulting to 5m")
		period = 5 * time.Minute
	}
	ticker := time.NewTicker(period)
	defer ticker.Stop()
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGTERM, syscall.SIGINT)
	for {
		select {
		case <-ctx.Done():
			return
		case <-signalChannel:
			return
		case <-ticker.C:
			client, err := newHttpClient(ctx, saJsonKeyPath)
			if err != nil {
				logger.Error(err, "error renewing GCP HTTP client")
			} else {
				clientMutex.Lock()
				gcpClient = client
				clientMutex.Unlock()
				logger.Info("GCP HTTP client renewed")
			}
		}
	}
}

func newHttpClient(ctx context.Context, saJsonKeyPath string) (*http.Client, error) {
	client, _, err := transport.NewHTTPClient(ctx, option.WithCredentialsFile(saJsonKeyPath), option.WithScopes("https://www.googleapis.com/auth/compute", "https://www.googleapis.com/auth/cloud-platform"))
	if err != nil {
		return nil, fmt.Errorf("error obtaining GCP HTTP client: [%w]", err)
	}
	return client, nil
}
