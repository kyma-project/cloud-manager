package looper

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

type SkrStatusHandleChecker interface {
	Check(t *testing.T) *KindHandleWrapper

	String() string
	AllHandlesChecked(t *testing.T)
	Len() int
	Handles() []*KindHandleWrapper
	Unchecked() SkrStatusHandleChecker

	Name(v string) SkrStatusHandleChecker
	Obj(v string) SkrStatusHandleChecker
	Crd(v string) SkrStatusHandleChecker
	Busola(v string) SkrStatusHandleChecker
	InstallerManifest() SkrStatusHandleChecker
	Indexer() SkrStatusHandleChecker
	Controller() SkrStatusHandleChecker
}

func NewSkrStatusChecker(status *SkrStatus) SkrStatusHandleChecker {
	var wraps []*KindHandleWrapper
	for i, h := range status.handles {
		wraps = append(wraps, &KindHandleWrapper{
			KindHandle: h,
			Index:      i,
		})
	}
	return &skrStatusHandleCheckerImpl{handles: wraps}
}

var _ SkrStatusHandleChecker = &skrStatusHandleCheckerImpl{}

type KindHandleWrapper struct {
	*KindHandle
	Index int
}

func (w *KindHandleWrapper) String() string {
	return fmt.Sprintf("%d %s", w.Index, w.KindHandle.String())
}

type skrStatusHandleCheckerImpl struct {
	handles []*KindHandleWrapper

	nok bool

	title string

	obj    string
	crd    string
	busola string

	objName string

	checkedIndexes map[int]struct{}
}

func (c *skrStatusHandleCheckerImpl) Len() int {
	return len(c.handles)
}

func (c *skrStatusHandleCheckerImpl) Handles() []*KindHandleWrapper {
	return c.handles
}

func (c *skrStatusHandleCheckerImpl) String() string {
	var arr []string
	for _, h := range c.handles {
		arr = append(arr, h.String())
	}
	return strings.Join(arr, "\n")
}

func (c *skrStatusHandleCheckerImpl) Unchecked() SkrStatusHandleChecker {
	if len(c.handles) == len(c.checkedIndexes) {
		return &skrStatusHandleCheckerImpl{handles: nil}
	}

	data := map[int]struct{}{}
	for i := 0; i < len(c.handles); i++ {
		data[i] = struct{}{}
	}
	for i := range c.checkedIndexes {
		delete(data, i)
	}

	var unchecked []*KindHandleWrapper
	for k := range data {
		unchecked = append(unchecked, c.handles[k])
	}

	return &skrStatusHandleCheckerImpl{handles: unchecked}
}

func (c *skrStatusHandleCheckerImpl) AllHandlesChecked(t *testing.T) {
	unchecked := c.Unchecked()
	if unchecked.Len() == 0 {
		return
	}
	var arr []string
	for _, h := range unchecked.Handles() {
		arr = append(arr, h.String())
	}
	assert.Fail(t, "Unchecked handles:\n"+strings.Join(arr, "\n"))
}

func (c *skrStatusHandleCheckerImpl) Check(t *testing.T) *KindHandleWrapper {
	if c.checkedIndexes == nil {
		c.checkedIndexes = map[int]struct{}{}
	}
	var handle *KindHandleWrapper
	var index int
	for i, h := range c.handles {
		if c.title != h.title {
			continue
		}
		if c.obj != h.objKindGroup {
			continue
		}
		if c.crd != h.crdKindGroup {
			continue
		}
		if c.busola != h.busolaKindGroup {
			continue
		}
		if c.objName != "" && c.objName != h.objName {
			continue
		}
		handle = h
		index = i
		break
	}
	if handle == nil {
		id := fmt.Sprintf("%s obj:%s crd:%s busola:%s name:%s", c.title, c.obj, c.crd, c.busola, c.objName)
		assert.Fail(t, "Unmatched handle: "+id)
		return nil
	}
	if c.nok && handle.ok {
		assert.Fail(t, fmt.Sprintf("Expected failed handle, but it's ok: %s", handle.String()))
	}
	if c.nok && !handle.ok {
		assert.Fail(t, fmt.Sprintf("Expected ok handle, but it's failed: %s", handle.String()))
	}
	if _, ok := c.checkedIndexes[index]; ok {
		assert.Fail(t, fmt.Sprintf("Handle already checked: %s", handle.String()))
	}
	c.checkedIndexes[index] = struct{}{}
	return handle
}

func (c *skrStatusHandleCheckerImpl) Name(v string) SkrStatusHandleChecker {
	c.objName = v
	return c
}

func (c *skrStatusHandleCheckerImpl) Obj(v string) SkrStatusHandleChecker {
	c.obj = v
	return c
}

func (c *skrStatusHandleCheckerImpl) Crd(v string) SkrStatusHandleChecker {
	c.obj = "customresourcedefinition.apiextensions.k8s.io"
	c.crd = v
	c.busola = ""
	return c
}

func (c *skrStatusHandleCheckerImpl) Busola(v string) SkrStatusHandleChecker {
	c.obj = "configmap"
	c.busola = v
	c.crd = ""
	return c
}

func (c *skrStatusHandleCheckerImpl) InstallerManifest() SkrStatusHandleChecker {
	c.title = "InstallerManifest"
	return c
}

func (c *skrStatusHandleCheckerImpl) Indexer() SkrStatusHandleChecker {
	c.title = "Indexer"
	return c
}

func (c *skrStatusHandleCheckerImpl) Controller() SkrStatusHandleChecker {
	c.title = "Controller"
	return c
}
