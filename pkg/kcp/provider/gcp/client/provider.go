package client

import (
	"context"
	"errors"
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/metrics"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/oauth2/v2"
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

type ClientProvider[T any] func(ctx context.Context, credentialsFile string) (T, error)

func NewCachedClientProvider[T comparable](p ClientProvider[T]) ClientProvider[T] {
	var result T
	var nilT T
	var m sync.Mutex
	return func(ctx context.Context, credentialsFile string) (T, error) {
		m.Lock()
		defer m.Unlock()
		var err error
		if result == nilT {
			result, err = p(ctx, credentialsFile)
		}
		return result, err
	}
}

var gcpClient *http.Client
var clientMutex sync.Mutex

func newCachedGcpClient(ctx context.Context, credentialsFile string) (*http.Client, error) {
	clientMutex.Lock()
	defer clientMutex.Unlock()
	if gcpClient == nil {
		client, err := newHttpClient(ctx, credentialsFile)
		if err != nil {
			return nil, err
		}
		gcpClient = client
		go renewCachedHttpClientPeriodically(context.Background(), credentialsFile, os.Getenv("GCP_CLIENT_RENEW_DURATION"))
	}
	return gcpClient, nil
}

func GetCachedGcpClient(ctx context.Context, credentialsFile string) (*http.Client, error) {
	if gcpClient == nil {
		return newCachedGcpClient(ctx, credentialsFile)
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

func renewCachedHttpClientPeriodically(ctx context.Context, credentialsFile, duration string) {
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
			client, err := newHttpClient(ctx, credentialsFile)
			if err != nil {
				logger.Error(err, "error renewing GCP HTTP client")
			} else {
				clientMutex.Lock()
				*gcpClient = *client
				clientMutex.Unlock()
				logger.Info("GCP HTTP client renewed")
			}
		}
	}
}

func newHttpClient(ctx context.Context, credentialsFile string) (*http.Client, error) {
	client, _, err := transport.NewHTTPClient(ctx, option.WithCredentialsFile(credentialsFile), option.WithScopes("https://www.googleapis.com/auth/compute", "https://www.googleapis.com/auth/cloud-platform"))
	if err != nil {
		return nil, fmt.Errorf("error obtaining GCP HTTP client: [%w]", err)
	}
	CheckGcpAuthentication(ctx, credentialsFile)
	client.Timeout = GcpConfig.GcpApiTimeout
	return client, nil
}

func CheckGcpAuthentication(ctx context.Context, credentialsFile string) {
	logger := composed.LoggerFromCtx(ctx)

	svc, err := oauth2.NewService(ctx, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		IncrementCallCounter("Authentication", "Check", "", err)
		logger.Error(err, "GCP Authentication Check - error creating new oauth2.Service")
		return
	}

	userInfoSvc := oauth2.NewUserinfoV2MeService(svc)
	userInfo, err := userInfoSvc.Get().Do()

	IncrementCallCounter("Authentication", "Check", "", err)
	if err != nil {
		logger.Error(err, "GCP Authentication Check - error getting UserInfo")
		return
	}

	logger.Info(fmt.Sprintf("GCP Authentication Check - successful [user = %s].", userInfo.Name))
}

func IncrementCallCounter(serviceName, operationName, region string, err error) {
	responseCode := 200
	if err != nil {
		var e *googleapi.Error
		if ok := errors.As(err, &e); ok {
			responseCode = e.Code
		} else {
			responseCode = 500
		}
	}
	gcpProject := ""
	metrics.CloudProviderCallCount.WithLabelValues(metrics.CloudProviderGCP, fmt.Sprintf("%s/%s", serviceName, operationName), fmt.Sprintf("%d", responseCode), region, gcpProject).Inc()
}
