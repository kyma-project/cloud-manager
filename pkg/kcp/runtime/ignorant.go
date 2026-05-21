package runtime

import (
	"github.com/kyma-project/cloud-manager/pkg/common/ignorant"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

var Ignore = ignorant.New()

var Tracker = composed.NewSimpleTracker(0, true)
