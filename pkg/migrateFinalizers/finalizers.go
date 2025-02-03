package migrateFinalizers

import "github.com/kyma-project/cloud-manager/api"

const newFinalizer = api.CommonFinalizerDeletionHook
const oldFinalizer1 = api.DO_NOT_USE_OLD_KcpFinalizer
const oldFinalizer2 = api.DO_NOT_USE_OLD_SkrFinalizer

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
