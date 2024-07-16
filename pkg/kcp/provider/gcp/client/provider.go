package client

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/metrics"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jws"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	"google.golang.org/api/transport"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"strings"
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
	CheckGcpAuthentication(ctx, client, saJsonKeyPath)
	return client, nil
}

func CheckGcpAuthentication(ctx context.Context, client *http.Client, saJsonKeyPath string) {
	logger := composed.LoggerFromCtx(ctx)
	jwtToken, err := GetJwtToken(ctx, saJsonKeyPath, 5*time.Minute, google.Endpoint.TokenURL, "openid")
	if err != nil {
		IncrementCallCounter("Authentication", "Check", "", err)
		logger.Error(err, "GCP Authentication Check - error getting JWT token")
		return
	}

	//Create the request body
	body := fmt.Sprintf("grant_type=%s&assertion=%s",
		url.QueryEscape("urn:ietf:params:oauth:grant-type:jwt-bearer"), jwtToken)
	reader := bytes.NewReader([]byte(body))

	//Perform the authentication check
	resp, err := client.Post(google.Endpoint.TokenURL, "application/x-www-form-urlencoded", reader)
	if err == nil {
		err = googleapi.CheckResponse(resp)
	}

	IncrementCallCounter("Authentication", "Check", "", err)
	if err != nil {
		logger.Error(err, "GCP Authentication Check - error performing authentication check")
		return
	}
	logger.Info("GCP Authentication Check - successful")
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

func GetJwtToken(ctx context.Context, saJsonKeyPath string, validity time.Duration, audience string, scopes ...string) (string, error) {

	//Get the credentials from the JSON Key file
	cred, err := transport.Creds(ctx, option.WithCredentialsFile(saJsonKeyPath))
	if err != nil {
		return "", err
	}

	//Parse the contents into a JWT Config
	jwtCfg, err := google.JWTConfigFromJSON(cred.JSON, scopes...)
	if err != nil {
		return "", err
	}
	jwtCfg.Audience = audience

	//Parse the Private Key
	pk, err := ParseKey(jwtCfg.PrivateKey)
	if err != nil {
		return "", err
	}

	//Get issue and expiry times
	now := time.Now()
	exp := now.Add(validity)

	//Create the JWT Claim Set
	cs := &jws.ClaimSet{
		Iss:   jwtCfg.Email,
		Sub:   jwtCfg.Email,
		Aud:   jwtCfg.Audience,
		Scope: strings.Join(jwtCfg.Scopes, " "),
		Iat:   now.Unix(),
		Exp:   exp.Unix(),
	}

	//Create the JWT Header
	hdr := &jws.Header{
		Algorithm: "RS256",
		Typ:       "JWT",
		KeyID:     string(jwtCfg.PrivateKeyID),
	}

	//Generate the JWT Token
	token, err := jws.Encode(hdr, cs, pk)
	if err != nil {
		return "", fmt.Errorf("error encoding jws in GetJwtToken: %w", err)
	}
	return token, nil
}

func ParseKey(key []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(key)
	if block != nil {
		key = block.Bytes
	}
	parsedKey, err := x509.ParsePKCS8PrivateKey(key)
	if err != nil {
		parsedKey, err = x509.ParsePKCS1PrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("private key should be a PEM or plain PKCS1 or PKCS8; parse error: %v", err)
		}
	}
	parsed, ok := parsedKey.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("private key is invalid")
	}
	return parsed, nil
}
