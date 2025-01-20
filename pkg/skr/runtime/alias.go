package runtime

import (
	"github.com/kyma-project/cloud-manager/pkg/skr/runtime/looper"
	"github.com/kyma-project/cloud-manager/pkg/skr/runtime/reconcile"
	"github.com/kyma-project/cloud-manager/pkg/skr/runtime/registry"
)

var NewLooper = looper.New
var NewActiveSkrCollection = looper.NewActiveSkrCollection

type SkrLooper = looper.SkrLooper

var NewRegistry = registry.New

type ReconcilerFactory = reconcile.ReconcilerFactory

type ReconcilerArguments = reconcile.ReconcilerArguments

var NewRunner = looper.NewSkrRunner
var NewSkrRunnerWithNoopStatusSaver = looper.NewSkrRunnerWithNoopStatusSaver

type SkrRegistry = registry.SkrRegistry

type SkrRunner = looper.SkrRunner

type ActiveSkrCollection = looper.ActiveSkrCollection
