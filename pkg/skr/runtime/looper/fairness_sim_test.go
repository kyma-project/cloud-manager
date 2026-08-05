//go:build looper_sim

package looper

import (
	"flag"
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/utils/clock"
)

// Dev-time fairness simulation. NOT part of the normal test run — it is guarded by the
// `looper_sim` build tag so `make test` / CI never compile or run it.
//
// Run:
//   go test -tags looper_sim -run TestLooperFairnessSim -v -timeout 30m \
//       ./pkg/skr/runtime/looper/ \
//       -args -sim.fleet=730 -sim.cyclic=24 -sim.notif=8 -sim.hot=3 \
//             -sim.connect=10ms -sim.duration=30s -sim.seed=1
//
// It exercises the REAL shipped fairness code (Queue, SkrGate, processOne, cyclicReAdd,
// Notify, recordNotifConnect) with a dummy sleeping handler in place of the per-SKR
// manager — no IO, no real 10s. Time is SCALED: what matters for fairness is the ratios
// (hot vs cold notification rate, connect time vs notifMinInterval, fleet/workers), not the
// absolute magnitudes. A ~30s wall-time run at these scaled values reproduces many hours of
// production rotation behavior.
//
// The report prints the theoretical fair mean gap vs the observed mean/percentile/max gaps
// and the worst offenders (the long tail). It fails if any SKR's max gap exceeds
// maxGapFactor × the mean — starvation is 5–20× the mean, so a generous factor cleanly
// separates the fairness bug from ordinary scheduling jitter.

var (
	simFleet    = flag.Int("sim.fleet", 730, "number of active SKRs")
	simCyclic   = flag.Int("sim.cyclic", 24, "cyclic worker count")
	simNotif    = flag.Int("sim.notif", 8, "notification worker count")
	simHot      = flag.Int("sim.hot", 3, "number of hyper-active (hot) SKRs")
	simConnect  = flag.Duration("sim.connect", 10*time.Millisecond, "simulated per-connect duration (scaled 10s)")
	simNotifMin = flag.Duration("sim.notifMin", 10*time.Millisecond, "per-SKR notification rate-limit interval (scaled)")
	simHotEvery = flag.Duration("sim.hotEvery", 2*time.Millisecond, "hot SKR notification firing period")
	simDuration = flag.Duration("sim.duration", 30*time.Second, "total simulated wall-clock run time")
	simSeed     = flag.Int64("sim.seed", 1, "RNG seed (fixed for reproducibility)")

	maxGapFactor = 3.0 // fail if any max gap > maxGapFactor × mean gap
)

// servedRecorder tracks per-SKR service timestamps and derives inter-service gaps.
type servedRecorder struct {
	mu     sync.Mutex
	last   map[string]time.Time
	gaps   map[string][]time.Duration
	counts map[string]int
}

func newServedRecorder() *servedRecorder {
	return &servedRecorder{
		last:   map[string]time.Time{},
		gaps:   map[string][]time.Duration{},
		counts: map[string]int{},
	}
}

func (r *servedRecorder) record(kyma string, now time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if prev, ok := r.last[kyma]; ok {
		r.gaps[kyma] = append(r.gaps[kyma], now.Sub(prev))
	}
	r.last[kyma] = now
	r.counts[kyma]++
}

func TestLooperFairnessSim(t *testing.T) {
	flag.Parse()
	rng := rand.New(rand.NewSource(*simSeed))

	// Real collection + looper wired with a dummy sleeping handler. The rate limiter uses
	// the real clock so the notification firing (real goroutines) is throttled realistically.
	col := newActiveSkrCollectionWithClock(clock.RealClock{})
	col.notifMinInterval = *simNotifMin

	rec := newServedRecorder()
	handle := func(_ int, kyma string) {
		time.Sleep(*simConnect)      // stand-in for CreateManager + 10s skrManager.Start
		rec.record(kyma, time.Now()) // record at end of connect (when the SKR is "served")
	}

	l := &skrLooper{
		ActiveSkrCollectionAdmin: col,
		logger:                   logr.Discard(),
		handleFn:                 handle,
		notificationConcurrency:  *simNotif,
		cyclicConcurrency:        *simCyclic,
		cyclicMinInterval:        60 * time.Second,
		reconcileTimeout:         10 * time.Second,
		cyclicImmediateThreshold: 1, // large fleet → FIFO re-add (matches prod)
	}

	// Seed the fleet: designate the first simHot as hot SKRs. Add to the cyclic queue,
	// which records membership (so Contains/Notify accept them) and primes the rotation.
	kymas := make([]string, *simFleet)
	for i := range kymas {
		kymas[i] = fmt.Sprintf("kyma-%04d", i)
	}
	hot := kymas[:*simHot]

	// Start the real workers (mirrors skrLooper.Start's worker launch).
	stop := make(chan struct{})
	var wg sync.WaitGroup
	l.ctx = t.Context() // handleFn ignores it; workers only need a live looper
	for x := 0; x < *simCyclic; x++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for !l.processOne(id, l.CyclicQueue(), "cyclic", l.cyclicReAdd, l.CyclicQueue().Add) {
			}
		}(x)
	}
	for x := 0; x < *simNotif; x++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for !l.processOne(id, l.NotificationQueue(), "notification", l.recordNotifConnect, func(string) {}) {
			}
		}(x)
	}

	// Prime the cyclic rotation (also records membership).
	for _, k := range kymas {
		l.CyclicQueue().Add(k)
	}

	// Notification driver: hot SKRs fire frequently; cold SKRs fire rarely (random).
	var drivers sync.WaitGroup
	drivers.Add(1)
	go func() {
		defer drivers.Done()
		ticker := time.NewTicker(*simHotEvery)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				for _, h := range hot {
					col.Notify(h)
				}
				// occasional cold-SKR notification (~1 per 50 ticks)
				if rng.Intn(50) == 0 {
					col.Notify(kymas[rng.Intn(len(kymas))])
				}
			}
		}
	}()

	time.Sleep(*simDuration)

	// Stop the drivers, then drain the queues so workers exit.
	close(stop)
	drivers.Wait()
	l.CyclicQueue().ShutDown()
	l.NotificationQueue().ShutDown()
	wg.Wait()

	report(t, rec, kymas, hot)
}

func report(t *testing.T, rec *servedRecorder, kymas, hot []string) {
	rec.mu.Lock()
	defer rec.mu.Unlock()

	// Flatten all gaps for global percentiles; track per-SKR max.
	var all []time.Duration
	maxGap := map[string]time.Duration{}
	var served int
	for _, k := range kymas {
		gs := rec.gaps[k]
		if len(gs) == 0 {
			continue
		}
		served++
		for _, g := range gs {
			all = append(all, g)
			if g > maxGap[k] {
				maxGap[k] = g
			}
		}
	}
	if len(all) == 0 {
		t.Fatal("no SKR was served — check sim wiring")
	}
	sort.Slice(all, func(i, j int) bool { return all[i] < all[j] })

	var sum time.Duration
	for _, g := range all {
		sum += g
	}
	mean := sum / time.Duration(len(all))
	pct := func(p float64) time.Duration { return all[int(float64(len(all)-1)*p)] }

	// Theoretical fair mean = fleet × connect / cyclicWorkers.
	fairMean := time.Duration(int64(*simFleet)*int64(*simConnect)) / time.Duration(*simCyclic)

	// Worst offenders by max gap.
	type kv struct {
		k string
		g time.Duration
	}
	worst := make([]kv, 0, len(maxGap))
	for k, g := range maxGap {
		worst = append(worst, kv{k, g})
	}
	sort.Slice(worst, func(i, j int) bool { return worst[i].g > worst[j].g })

	t.Logf("=== SKR Looper fairness simulation report ===")
	t.Logf("fleet=%d cyclic=%d notif=%d hot=%d connect=%s notifMin=%s duration=%s seed=%d",
		*simFleet, *simCyclic, *simNotif, *simHot, *simConnect, *simNotifMin, *simDuration, *simSeed)
	t.Logf("served %d/%d SKRs at least twice (need 2 services to measure a gap)", served, len(kymas))
	t.Logf("theoretical fair mean gap = %s", fairMean)
	t.Logf("observed gap: mean=%s p50=%s p95=%s p99=%s max=%s",
		mean, pct(0.50), pct(0.95), pct(0.99), all[len(all)-1])

	t.Logf("worst 10 SKRs by max gap (× mean):")
	for i := 0; i < 10 && i < len(worst); i++ {
		ratio := float64(worst[i].g) / float64(mean)
		hotMark := ""
		for _, h := range hot {
			if h == worst[i].k {
				hotMark = " [HOT]"
			}
		}
		t.Logf("  %s  max=%s  %.1f×mean%s", worst[i].k, worst[i].g, ratio, hotMark)
	}

	// Hot SKRs should be served frequently (bonus connects) and NOT be in the starved tail.
	t.Logf("hot SKR connect counts:")
	for _, h := range hot {
		t.Logf("  %s  connects=%d", h, rec.counts[h])
	}

	// Fairness assertion: no SKR's max gap may exceed maxGapFactor × mean.
	worstRatio := float64(worst[0].g) / float64(mean)
	if worstRatio > maxGapFactor {
		t.Errorf("FAIRNESS FAIL: worst max gap %s is %.1f× mean (%s), exceeds %.1f× — long tail / starvation present",
			worst[0].g, worstRatio, mean, maxGapFactor)
	} else {
		t.Logf("FAIRNESS OK: worst max gap %.1f× mean (limit %.1f×)", worstRatio, maxGapFactor)
	}
}
