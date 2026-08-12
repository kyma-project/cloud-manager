package keb

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/elliotchance/pie/v2"
	"github.com/go-logr/logr"
	"github.com/hashicorp/go-multierror"
	"github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"
	"github.com/kyma-project/cloud-manager/pkg/util"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
)

// WaitOption ============================================================

type WaitOption interface {
	ApplyOnWait(*waitOptions)
}

type waitOptions struct {
	runtimeId             string
	alias                 string
	timeout               time.Duration
	interval              time.Duration
	progressCallback      func(WaitProgress)
	logger                logr.Logger
	errorDuration         time.Duration
	terminalErrorDuration time.Duration
	terminalRetryLimit    int
	sleeper               util.Sleeper
	nowFunc               func() time.Time
}

func (o *waitOptions) validate() error {
	var result error
	if o.runtimeId == "" && o.alias == "" {
		result = multierror.Append(result, errors.New("no runtimeId/alias specified to wait for"))
	}
	if result != nil {
		return fmt.Errorf("waitCompleted invalid input: %w", result)
	}
	if o.timeout == 0 {
		o.timeout = 15 * time.Minute
	}
	if o.interval == 0 {
		o.interval = 10 * time.Second
	}
	if o.errorDuration == 0 {
		o.errorDuration = 10 * time.Minute
	}
	if o.terminalErrorDuration == 0 {
		o.terminalErrorDuration = 5 * time.Minute
	}
	if o.terminalRetryLimit == 0 {
		o.terminalRetryLimit = 3
	}
	if o.sleeper == nil {
		o.sleeper = util.SleeperFunc(util.RealSleeperFunc)
	}
	if o.nowFunc == nil {
		o.nowFunc = time.Now
	}
	return nil
}

type WaitProgress struct {
	Done    []InstanceDetails
	Pending []InstanceDetails
	WithErr []InstanceDetails
	Changed bool
}

func (in WaitProgress) DoneAliases() []string {
	return pie.Map(in.Done, func(x InstanceDetails) string {
		return x.Alias
	})
}

func (in WaitProgress) PendingAliases() []string {
	return pie.Map(in.Pending, func(x InstanceDetails) string {
		return x.Alias
	})
}

func (in WaitProgress) ErrAliases() []string {
	return pie.Map(in.WithErr, func(x InstanceDetails) string {
		return x.Alias
	})
}

func (in WaitProgress) Hash() string {
	arr := make([]string, 0, len(in.Done)+1+len(in.Pending))
	for _, i := range in.Done {
		arr = append(arr, i.RuntimeID)
	}
	arr = append(arr, "|")
	for _, i := range in.Pending {
		arr = append(arr, i.RuntimeID)
	}
	txt := strings.Join(arr, ",")
	hasher := sha256.New()
	hasher.Write([]byte(txt))
	sum := hasher.Sum(nil)
	return hex.EncodeToString(sum)
}

var defaultWaitOptions = []WaitOption{
	WithTimeout(15 * time.Minute),
	WithInterval(5 * time.Second),
	WithProgressCallback(func(WaitProgress) {}),
	WithErrorDuration(10 * time.Minute),
	WithTerminalErrorDuration(5 * time.Minute),
	WithSleeperFunc(util.RealSleeperFunc),
}

// WaitHandler combines listing instances with the ability to force-retry a failed shoot.
type WaitHandler interface {
	InstanceLister
	ShootRetrier
}

func WaitCompleted(ctx context.Context, handler WaitHandler, opts ...WaitOption) error {
	options := &waitOptions{}
	for _, o := range append(append([]WaitOption{}, defaultWaitOptions...), opts...) {
		o.ApplyOnWait(options)
	}
	if err := options.validate(); err != nil {
		return err
	}

	lastNotifyHash := "-"

	// tracks the first time a transient error was seen per runtimeID
	errorFirstSeen := map[string]time.Time{}

	// tracks terminal shoot errors: first detection, retry count, and last retry time
	terminalErrorFirstSeen := map[string]time.Time{}
	terminalRetryCount := map[string]int{}
	lastRetryAt := map[string]time.Time{}

	// map wait options to list options
	var listOpts []ListOption
	for _, o := range opts {
		if lo, ok := o.(ListOption); ok {
			listOpts = append(listOpts, lo)
		}
	}

	var loopErr error
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, options.timeout)
	defer cancel()

outerLoop:
	for {
		arr, err := handler.List(ctx, listOpts...)
		if apierrors.IsNotFound(err) || meta.IsNoMatchError(err) {
			break
		}
		if err != nil {
			loopErr = fmt.Errorf("error listing instances in WaitCompleted: %w", err)
			break
		}

		if arr == nil {
			arr = []InstanceDetails{}
		}
		b, err := json.Marshal(arr)
		if err == nil {
			options.logger.WithValues("instances", string(b)).Info("WaitCompleted poll status")
		}

		var done []InstanceDetails
		var pending []InstanceDetails
		var withErr []InstanceDetails

		for _, id := range arr {
			if id.BeingDeleted {
				id.Message = "being deleted"
				pending = append(pending, id)
			} else if id.State == infrastructuremanagerv1.RuntimeStateFailed {
				withErr = append(withErr, id)
			} else if id.State == infrastructuremanagerv1.RuntimeStatePending {
				pending = append(pending, id)
			} else if id.ProvisioningCompleted {
				done = append(done, id)
			} else {
				pending = append(pending, id)
			}
		}

		wp := WaitProgress{
			Done:    done,
			Pending: pending,
			WithErr: withErr,
		}

		currentNotifyHash := wp.Hash()
		wp.Changed = currentNotifyHash != lastNotifyHash
		lastNotifyHash = currentNotifyHash
		options.progressCallback(wp)

		// process instances in error state
		for _, id := range withErr {
			if id.HasTerminalShootError() {
				// record first detection
				if _, seen := terminalErrorFirstSeen[id.RuntimeID]; !seen {
					terminalErrorFirstSeen[id.RuntimeID] = options.nowFunc()
				}
				// bounded, debounced retry
				if terminalRetryCount[id.RuntimeID] < options.terminalRetryLimit {
					last, retried := lastRetryAt[id.RuntimeID]
					if !retried || options.nowFunc().Sub(last) >= options.interval {
						if retryErr := handler.ForceShootRetry(ctx, id.RuntimeID); retryErr != nil {
							options.logger.Error(retryErr, "ForceShootRetry failed", "runtimeID", id.RuntimeID)
						}
						terminalRetryCount[id.RuntimeID]++
						lastRetryAt[id.RuntimeID] = options.nowFunc()
						options.logger.Info("forced shoot retry", "runtimeID", id.RuntimeID, "attempt", terminalRetryCount[id.RuntimeID])
					}
				}
				// fail once the tolerance window is exhausted
				if options.nowFunc().Sub(terminalErrorFirstSeen[id.RuntimeID]) > options.terminalErrorDuration {
					loopErr = fmt.Errorf("instance %s %s has terminal shoot error: %q\n%s", id.Alias, id.RuntimeID, id.Message, id.ShootInfo())
					break outerLoop
				}
				continue
			}
			// transient error: record when we first saw it
			if _, seen := errorFirstSeen[id.RuntimeID]; !seen {
				errorFirstSeen[id.RuntimeID] = options.nowFunc()
			}
		}

		// clear recovered instances from all error trackers
		withErrSet := make(map[string]bool, len(withErr))
		for _, id := range withErr {
			withErrSet[id.RuntimeID] = true
		}
		for runtimeID := range errorFirstSeen {
			if !withErrSet[runtimeID] {
				delete(errorFirstSeen, runtimeID)
			}
		}
		for runtimeID := range terminalErrorFirstSeen {
			if !withErrSet[runtimeID] {
				delete(terminalErrorFirstSeen, runtimeID)
				delete(terminalRetryCount, runtimeID)
				delete(lastRetryAt, runtimeID)
			}
		}

		// check if any transient error has exceeded the tolerance window
		err = nil
		for runtimeID, since := range errorFirstSeen {
			if options.nowFunc().Sub(since) > options.errorDuration {
				var id *InstanceDetails
				for _, x := range arr {
					if x.RuntimeID == runtimeID {
						xx := x
						id = new(xx)
						break
					}
				}
				if id != nil {
					err = multierror.Append(err, fmt.Errorf("instance %s %s has error %q\n%s", id.Alias, id.RuntimeID, id.Message, id.ShootInfo()))
				}
			}
		}

		if err != nil {
			loopErr = err
			break
		}
		// early exit when nothing left to wait for
		if len(pending) == 0 && len(withErr) == 0 {
			break
		}

		options.sleeper.Sleep(ctx, options.interval)
	}

	if loopErr != nil {
		return fmt.Errorf("error waiting for instance(s) to become provisioned: %w", loopErr)
	}

	return nil
}
