package runtime

import (
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/skr/runtime/looper"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/skr/runtime/registry"
)

var NewLooper = looper.New

var NewRegistry = registry.New

var NewRunner = looper.NewSkrRunner

type SkrRegistry = registry.SkrRegistry

type SkrRunner = looper.SkrRunner

type ActiveSkrCollection = looper.ActiveSkrCollection
