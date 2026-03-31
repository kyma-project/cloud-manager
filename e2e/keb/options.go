package keb

import (
	"fmt"
	"time"

	"github.com/go-logr/logr"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	e2elib "github.com/kyma-project/cloud-manager/e2e/lib"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

// WithNodesRange ======================================================================

// WithNodesRange specifies nodes rage for create
type WithNodesRange string

func (o WithNodesRange) ApplyOnCreate(opts *createOptions) {
	opts.nodesRange = string(o)
}

// WithPodsRange ========================================================================

// WithPodsRange specifies pods range for create
type WithPodsRange string

func (o WithPodsRange) ApplyOnCreate(opts *createOptions) {
	opts.podsRange = string(o)
}

// WithServicesRange ======================================================================

// WithServicesRange specifies services range for create
type WithServicesRange string

func (o WithServicesRange) ApplyOnCreate(opts *createOptions) {
	opts.servicesRange = string(o)
}

// WithRegion ============================================================================

type WithRegion string

func (o WithRegion) ApplyOnCreate(opt *createOptions) {
	opt.region = string(o)
}

// WithAlias ===============================================================================

// WithAlias specifies alias for list and create
type WithAlias string

func (o WithAlias) ApplyOnList(opt *listOptions) {
	opt.alias = string(o)
}

func (o WithAlias) ApplyOnCreate(opt *createOptions) {
	opt.alias = string(o)
}

func (o WithAlias) ApplyOnWait(opt *waitOptions) {
	opt.alias = string(o)
}

// WithGlobalAccount ====================================================================

// WithGlobalAccount specifies alias for list and create
type WithGlobalAccount string

func (o WithGlobalAccount) ApplyOnList(opt *listOptions) {
	opt.globalAccount = string(o)
}

func (o WithGlobalAccount) ApplyOnCreate(opt *createOptions) {
	opt.globalAccount = string(o)
}

// WithSubAccount =======================================================================

// WithSubAccount specifies alias for list and create
type WithSubAccount string

func (o WithSubAccount) ApplyOnList(opt *listOptions) {
	opt.globalAccount = string(o)
}

func (o WithSubAccount) ApplyOnCreate(opt *createOptions) {
	opt.subAccount = string(o)
}

// WithProvider ========================================================================

// WithProvider specifies alias for list and create
type WithProvider cloudcontrolv1beta1.ProviderType

func (o WithProvider) ApplyOnList(opt *listOptions) {
	opt.provider = cloudcontrolv1beta1.ProviderType(o)
}

func (o WithProvider) ApplyOnCreate(opt *createOptions) {
	opt.provider = cloudcontrolv1beta1.ProviderType(o)
	if opt.region == "" {
		opt.region = e2elib.DefaultRegions[opt.provider]
	}
}

// WithTimeout ===============================================================================

// WithTimeout specifies timeout for
// * wait ops WaitProvisioningCompleted and WaitDeleted for how much it wil be waited for instances to be provisioned or deleted
// * DeleteInstance for how much it will be waited for shoot to be marked as deleted, if zero no wait is done
// * CreateInstance for how much it will be waited for shoot to be created, if zero no wait is done
type WithTimeout time.Duration

func (o WithTimeout) ApplyOnDelete(opt *deleteOptions) {
	opt.shootMarkedForDeletionTimeout = time.Duration(o)
}

func (o WithTimeout) ApplyOnCreate(opt *createOptions) {
	opt.shootCreatedTimeout = time.Duration(o)
}

func (o WithTimeout) ApplyOnWait(opt *waitOptions) {
	opt.timeout = time.Duration(o)
}

// WithInterval ====================================================================================

// WithInterval specifies interval duration for timeout if any set, in wait ops, DeleteInstance and CreateInstance
type WithInterval time.Duration

func (o WithInterval) ApplyOnDelete(opt *deleteOptions) {
	opt.shootMarkedForDeletionInterval = time.Duration(o)
}

func (o WithInterval) ApplyOnCreate(opt *createOptions) {
	opt.shootCreatedInterval = time.Duration(o)
}

func (o WithInterval) ApplyOnWait(opt *waitOptions) {
	opt.interval = time.Duration(o)
}

// WithRuntime ====================================================================================

// WithRuntime specifies runtime id
type WithRuntime string

func (o WithRuntime) ApplyOnDelete(opt *deleteOptions) {
	opt.runtimeId = string(o)
}

func (o WithRuntime) ApplyOnList(opt *listOptions) {
	opt.runtimeId = string(o)
}

func (o WithRuntime) ApplyOnWait(opt *waitOptions) {
	opt.runtimeId = string(o)
}

// WithErrorThreshold ===========================================================================

type WithErrorCountThreshold int

func (o WithErrorCountThreshold) ApplyOnWait(opt *waitOptions) {
	opt.errorCountThreshold = int(o)
}

// WithSleeper ==================================================================================

type WithSleeperOpt struct {
	sleeper util.Sleeper
}

func (o WithSleeperOpt) ApplyOnWait(opt *waitOptions) {
	opt.sleeper = o.sleeper
}

func WithSleeper(sleeper util.Sleeper) WithSleeperOpt {
	return WithSleeperOpt{sleeper}
}

type WithSleeperFunc util.SleeperFunc

func (o WithSleeperFunc) ApplyOnWait(opt *waitOptions) {
	opt.sleeper = util.SleeperFunc(o)
}

// WithProgressCallback ============================================================================

// WithProgressCallback specifies progress callback for wait operations in WaitProvisioningCompleted
// and WaitDeleted
type WithProgressCallback func(WaitProgress)

func (o WithProgressCallback) ApplyOnWait(opt *waitOptions) {
	opt.progressCallback = o
}

func WaitProgressPrint() WithProgressCallback {
	lastProgressPrintTime := time.Now()
	return func(progress WaitProgress) {
		if progress.Changed || time.Since(lastProgressPrintTime) > time.Minute {
			lastProgressPrintTime = time.Now()
			fmt.Printf("%s\n", time.Now().Format(time.RFC3339))
			fmt.Printf("Pending: %v\n", progress.PendingAliases())
			fmt.Printf("Done: %v\n", progress.DoneAliases())
			fmt.Printf("WithErr: %v\n", progress.ErrAliases())
		}
	}
}

// WithLogger ============================================================================

type WithLogger logr.Logger

func (o WithLogger) ApplyOnCreate(opts *skrManagerFactoryCreateOptions) {
	l := logr.Logger(o)
	opts.logger = &l
}

func (o WithLogger) ApplyOnWait(opts *waitOptions) {
	opts.logger = logr.Logger(o)
}
