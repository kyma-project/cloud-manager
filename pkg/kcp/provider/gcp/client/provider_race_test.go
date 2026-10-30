package client

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Helpers below stand in for the renewal path in unit tests. The real renewal
// (renewCachedHttpClientPeriodically) can't run here because it builds a
// network-backed GCP client; these seams perform the same writes to the cache
// global and change together with the production fix.

// testSeedCachedClient seeds the cache so GetCachedGcpClient takes its fast
// read path without building a real client.
func testSeedCachedClient(c *http.Client) {
	gcpClient.Store(c)
}

// testStoreCachedClientPtr mirrors the mutex-guarded pointer assignment the
// cache performs when (re)populating the global client.
func testStoreCachedClientPtr(c *http.Client) {
	clientMutex.Lock()
	gcpClient.Store(c)
	clientMutex.Unlock()
}

// Cycle 1: concurrently resolving the cached client while the renewal path
// writes the global pointer must not race.
func TestGetCachedGcpClientConcurrentResolveAndRenewNoRace(t *testing.T) {
	testSeedCachedClient(&http.Client{Timeout: time.Second})

	var wg sync.WaitGroup
	stop := make(chan struct{})

	// renewal-shaped writer: reassign the global pointer under the mutex
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
				testStoreCachedClientPtr(&http.Client{Timeout: time.Second})
			}
		}
	}()

	// readers using the real production fast path
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 200 {
				c, err := GetCachedGcpClient(context.Background(), "")
				assert.NoError(t, err)
				assert.NotNil(t, c)
			}
		}()
	}

	// let readers run against the writer, then stop
	time.Sleep(50 * time.Millisecond)
	close(stop)
	wg.Wait()
}
