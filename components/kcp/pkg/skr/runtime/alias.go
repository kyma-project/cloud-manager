package runtime

import (
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/skr/runtime/looper"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/skr/runtime/registry"
)

var NewLooper = looper.New

var NewRegistry = registry.New

type SkrRegistry = registry.SkrRegistry

type ActiveSkrCollection = looper.ActiveSkrCollection
