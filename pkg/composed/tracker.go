package composed

import (
	"fmt"
	"strings"

	"github.com/hashicorp/golang-lru/v2"
)

type Tracker interface {
	Record(objName string, reconciliationLabel string)
	Get(name string) (string, bool)
	Clear(name string)
	IsReconciledWith(name, expectedReconciliationLabel string) error
}

func NewSimpleTracker(size int, nameOnly bool) Tracker {
	if size <= 0 {
		size = 128
	}
	cache, _ := lru.New[string, string](size)
	return &simpleTracker{
		cache:    cache,
		nameOnly: nameOnly,
	}
}

type simpleTracker struct {
	cache    *lru.Cache[string, string]
	nameOnly bool
}

func (t *simpleTracker) sanitizeName(name string) string {
	if !t.nameOnly {
		return name
	}
	chunks := strings.Split(name, "/")
	if len(chunks) > 1 {
		return chunks[1]
	}
	return chunks[0]
}

func (t *simpleTracker) Clear(name string) {
	name = t.sanitizeName(name)
	t.cache.Remove(name)
}

func (t *simpleTracker) Record(name string, reconciliationLabel string) {
	name = t.sanitizeName(name)
	t.cache.Add(name, reconciliationLabel)
}

func (t *simpleTracker) Get(name string) (string, bool) {
	name = t.sanitizeName(name)
	return t.cache.Get(name)
}

func (t *simpleTracker) IsReconciledWith(name, expectedReconciliationLabel string) error {
	actual, ok := t.cache.Get(name)
	if !ok {
		return fmt.Errorf("%s is not reconciled", name)
	}
	if actual != expectedReconciliationLabel {
		return fmt.Errorf("%s is not reconciled with %s but expected is %s", name, actual, expectedReconciliationLabel)
	}
	return nil
}
