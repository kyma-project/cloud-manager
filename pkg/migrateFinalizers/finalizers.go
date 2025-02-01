package migrateFinalizers

import "github.com/kyma-project/cloud-manager/api"

const newFinalizer = api.CommonFinalizerDeletionHook
const oldFinalizer1 = "cloud-control.kyma-project.io/deletion-hook"
const oldFinalizer2 = "cloud-resources.kyma-project.io/deletion-hook"

type finalizerInfo struct {
	AddFinalizers    []string
	RemoveFinalizers []string
}

func newFinalizerInfo() *finalizerInfo {
	return &finalizerInfo{
		AddFinalizers:    []string{newFinalizer},
		RemoveFinalizers: []string{oldFinalizer1, oldFinalizer2},
	}
}
