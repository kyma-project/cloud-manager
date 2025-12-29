package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/config"
	"github.com/kyma-project/cloud-manager/pkg/metrics"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
	"google.golang.org/api/transport"
	"sigs.k8s.io/controller-runtime/pkg/log"
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
		go renewCachedHttpClientPeriodically(context.Background(), credentialsFile, config.GcpConfig.ClientRenewDuration)
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

func renewCachedHttpClientPeriodically(ctx context.Context, credentialsFile string, d time.Duration) {
	logger := log.FromContext(ctx)
	ticker := time.NewTicker(d)
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

func loadCredentials(ctx context.Context, credentialsFile string, scopes ...string) (*google.Credentials, error) {
	credentialsData, err := os.ReadFile(credentialsFile)
	if err != nil {
		return nil, fmt.Errorf("error reading credentials file: [%w]", err)
	}
	credentials, err := google.CredentialsFromJSON(ctx, credentialsData, scopes...)
	if err != nil {
		return nil, fmt.Errorf("error loading credentials: [%w]", err)
	}
	return credentials, nil
}

func newHttpClient(ctx context.Context, credentialsFile string) (*http.Client, error) {
	credentials, err := loadCredentials(ctx, credentialsFile,
		"https://www.googleapis.com/auth/compute",
		"https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return nil, err
	}

	client, _, err := transport.NewHTTPClient(ctx, option.WithTokenSource(credentials.TokenSource))
	if err != nil {
		return nil, fmt.Errorf("error obtaining GCP HTTP client: [%w]", err)
	}
	CheckGcpAuthentication(ctx, credentialsFile)
	client.Timeout = config.GcpConfig.GcpApiTimeout
	return client, nil
}

func CheckGcpAuthentication(ctx context.Context, credentialsFile string) {
	logger := composed.LoggerFromCtx(ctx)

	credentials, err := loadCredentials(ctx, credentialsFile, oauth2.UserinfoEmailScope)
	if err != nil {
		IncrementCallCounter("Authentication", "Check", "", err)
		logger.Error(err, "GCP Authentication Check - failed to load credentials")
		return
	}

	svc, err := oauth2.NewService(ctx, option.WithTokenSource(credentials.TokenSource))
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
