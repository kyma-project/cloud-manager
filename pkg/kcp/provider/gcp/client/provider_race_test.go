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

// testSnapshotCachedClient restores the cache global after the test so these
// white-box tests don't leak a populated cache into later tests (which would
// silently push them onto the fast read path).
func testSnapshotCachedClient(t *testing.T) {
	saved := gcpClient.Load()
	t.Cleanup(func() { gcpClient.Store(saved) })
}

// testSeedCachedClient seeds the cache so GetCachedGcpClient takes its fast
// read path without building a real client.
func testSeedCachedClient(c *http.Client) {
	gcpClient.Store(c)
}

// testStoreCachedClientPtr mirrors the mutex-guarded store the cache performs
// when it first populates the global client (newCachedGcpClient).
func testStoreCachedClientPtr(c *http.Client) {
	clientMutex.Lock()
	gcpClient.Store(c)
	clientMutex.Unlock()
}

// testRenewCachedClient mirrors the bare atomic store renewCachedHttpClientPeriodically
// performs on a renewal tick (no mutex) to swap in the freshly built client.
func testRenewCachedClient(c *http.Client) {
	gcpClient.Store(c)
}

// Cycle 1: concurrently resolving the cached client while both cache-write
// paths run must not race.
//
// Only meaningful under `go test -race`: without the race detector the
// assertions below are vacuous (any seeded cache satisfies them) and the
// concurrent read/write defect goes undetected. CI runs this package with -race.
func TestGetCachedGcpClientConcurrentResolveAndRenewNoRace(t *testing.T) {
	testSnapshotCachedClient(t)
	testSeedCachedClient(&http.Client{Timeout: time.Second})

	const iterations = 200
	var wg sync.WaitGroup

	// writers: exercise both production write paths concurrently with readers —
	// the mutex-guarded initial population and the renewal's bare atomic store.
	wg.Add(2)
	go func() {
		defer wg.Done()
		for range iterations {
			testStoreCachedClientPtr(&http.Client{Timeout: time.Second})
		}
	}()
	go func() {
		defer wg.Done()
		for range iterations {
			testRenewCachedClient(&http.Client{Timeout: time.Second})
		}
	}()

	// readers using the real production fast path
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range iterations {
				c, err := GetCachedGcpClient(context.Background(), "")
				assert.NoError(t, err)
				assert.NotNil(t, c)
			}
		}()
	}

	wg.Wait()
}

// Cycle 2: a renewal must not mutate a client a caller already holds. The
// renewal swaps the cached pointer in, leaving previously-returned clients
// valid and unchanged.
func TestGetCachedGcpClientRenewalDoesNotMutateExistingReference(t *testing.T) {
	testSnapshotCachedClient(t)
	original := &http.Client{Timeout: 11 * time.Second}
	testSeedCachedClient(original)

	held, err := GetCachedGcpClient(context.Background(), "")
	assert.NoError(t, err)
	assert.Equal(t, 11*time.Second, held.Timeout)

	// a renewal installs a different client
	testRenewCachedClient(&http.Client{Timeout: 22 * time.Second})

	// the previously-returned client must be unchanged
	assert.Same(t, original, held)
	assert.Equal(t, 11*time.Second, held.Timeout)
}
