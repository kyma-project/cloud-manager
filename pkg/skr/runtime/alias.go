package runtime

import (
	"github.com/kyma-project/cloud-manager/pkg/skr/runtime/looper"
	"github.com/kyma-project/cloud-manager/pkg/skr/runtime/reconcile"
	"github.com/kyma-project/cloud-manager/pkg/skr/runtime/registry"
	"github.com/kyma-project/cloud-manager/pkg/skr/runtime/reload"
)

var NewLooper = looper.New

type SkrLooper = looper.SkrLooper

var NewRegistry = registry.New

type ReconcilerFactory = reconcile.ReconcilerFactory

type ReconcilerArguments = reconcile.ReconcilerArguments

var NewRunner = looper.NewSkrRunner

type SkrRegistry = registry.SkrRegistry

type Reloader = reload.Reloader

type SkrRunner = looper.SkrRunner

type ActiveSkrCollection = looper.ActiveSkrCollection
