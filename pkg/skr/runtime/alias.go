package runtime

import (
	"github.com/kyma-project/cloud-manager/pkg/skr/runtime/looper"
	"github.com/kyma-project/cloud-manager/pkg/skr/runtime/registry"
)

var NewLooper = looper.New

type SkrLooper = looper.SkrLooper

var NewRegistry = registry.New

type ReconcilerFactory = registry.ReconcilerFactory

var NewRunner = looper.NewSkrRunner

type SkrRegistry = registry.SkrRegistry

type SkrRunner = looper.SkrRunner

type ActiveSkrCollection = looper.ActiveSkrCollection
