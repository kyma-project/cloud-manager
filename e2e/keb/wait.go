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
	runtimeId           string
	alias               string
	timeout             time.Duration
	interval            time.Duration
	progressCallback    func(WaitProgress)
	logger              logr.Logger
	errorCountThreshold int
	sleeper             util.Sleeper
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
	if o.errorCountThreshold == 0 {
		o.errorCountThreshold = 12 // with interval 5s = 1min
	}
	if o.sleeper == nil {
		o.sleeper = util.SleeperFunc(util.RealSleeperFunc)
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
	WithErrorCountThreshold(12), // with interval 5s = 1min
	WithSleeperFunc(util.RealSleeperFunc),
}

func WaitCompleted(ctx context.Context, lister InstanceLister, opts ...WaitOption) error {
	options := &waitOptions{}
	for _, o := range append(append([]WaitOption{}, defaultWaitOptions...), opts...) {
		o.ApplyOnWait(options)
	}
	if err := options.validate(); err != nil {
		return err
	}

	lastNotifyHash := "-"

	runtimeErrorCount := map[string]int{}

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
	for {
		arr, err := lister.List(ctx, listOpts...)
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

		// increase the error count for this runtime
		for _, id := range withErr {
			v := runtimeErrorCount[id.RuntimeID]
			v++
			runtimeErrorCount[id.RuntimeID] = v
		}

		// go through all runtimes with error counts and make err with those crossing the threshold
		err = nil
		for runtimeID, errorCount := range runtimeErrorCount {
			if errorCount > options.errorCountThreshold {
				var id *InstanceDetails
				for _, x := range arr {
					if x.RuntimeID == runtimeID {
						xx := x
						id = &xx
						break
					}
				}
				if id != nil {
					err = multierror.Append(err, fmt.Errorf("instance %s %s has error %q", id.Alias, id.RuntimeID, id.Message))
				}
			}
		}

		if err != nil {
			loopErr = err
			break
		}
		// this is early exit, exiting the loop if no more pending, or there's some instance with error
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
