package client

import (
	"context"
	"google.golang.org/api/option"
	"google.golang.org/api/transport"
	"net/http"
	"sync"
)

type ClientProvider[T any] func(ctx context.Context, httpClient *http.Client) (T, error)

func NewCachedClientProvider[T comparable](p ClientProvider[T]) ClientProvider[T] {
	var result T
	var nilT T
	var m sync.Mutex
	return func(ctx context.Context, httpClient *http.Client) (T, error) {
		m.Lock()
		defer m.Unlock()
		var err error
		if result == nilT {
			result, err = p(ctx, gcpClient)
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
			return nil, err
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
