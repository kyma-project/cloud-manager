package looper

import (
	"sync"

	"github.com/kyma-project/cloud-manager/pkg/metrics"
)

// SkrGate guarantees at most one live manager per kymaName across all worker
// pools. TryClaim is a short O(1) critical section; the returned claim is held
// for the whole handleOneSkr lifetime (~10s) but the gate mutex is NOT held
// during that time.
type SkrGate struct {
	mu       sync.Mutex
	inFlight map[string]struct{}
}

func NewSkrGate() *SkrGate { return &SkrGate{inFlight: map[string]struct{}{}} }

// TryClaim atomically inserts kymaName. Returns true if the caller now owns the
// claim, false if another worker (either sleeve) already holds it. Never blocks.
func (g *SkrGate) TryClaim(kymaName string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	if _, held := g.inFlight[kymaName]; held {
		return false
	}
	g.inFlight[kymaName] = struct{}{}
	metrics.SkrLooperGateInFlight.Inc()
	return true
}

// Release removes the claim. Idempotent (defensive against double-release).
func (g *SkrGate) Release(kymaName string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if _, held := g.inFlight[kymaName]; held {
		delete(g.inFlight, kymaName)
		metrics.SkrLooperGateInFlight.Dec()
	}
}

// heldForTest reports whether a claim is currently held. Test-only helper.
func (g *SkrGate) heldForTest(kymaName string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	_, held := g.inFlight[kymaName]
	return held
}
