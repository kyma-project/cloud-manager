package looper

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type SkrStatusHandleChecker interface {
	CheckAll(t *testing.T, testCases []SkrStatusTestCase) SkrStatusHandleChecker
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

type SkrStatusTestCase struct {
	Kind     string
	Ok       bool
	Title    string
	Type     KindForm
	Outcomes []string
}

func (tc SkrStatusTestCase) String() string {
	return fmt.Sprintf("%s %v %s %s: %v", tc.Kind, tc.Ok, tc.Title, tc.Type, tc.Outcomes)
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

func (c *skrStatusHandleCheckerImpl) CheckAll(t *testing.T, testCases []SkrStatusTestCase) SkrStatusHandleChecker {
	for _, testCase := range testCases {
		switch testCase.Type {
		case KindFormObj:
			c.Obj(testCase.Kind)
		case KindFormCrd:
			c.Crd(testCase.Kind)
		case KindFormBusola:
			c.Busola(testCase.Kind)
		default:
			assert.Fail(t, fmt.Sprintf("Unknown kind in test case: %s", testCase.String()))
			return c
		}
		c.Ok(testCase.Ok)
		h := c.Check(t)
		if h != nil {
			if testCase.Title != "" {
				assert.Equal(t, testCase.Title, h.title)
			}
			if len(h.outcomes) > 0 {
				assert.Equal(t, testCase.Outcomes, h.outcomes)
			}
		}
	}
	return c
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
		assert.Fail(t, "Unmatched handle: "+id+" This kind was asserted for in the test, but it was not installed.")
		return nil
	}
	if c.nok && handle.ok {
		assert.Fail(t, fmt.Sprintf("Expected failed handle, but it's ok: %s", handle.String()))
	}
	if !c.nok && !handle.ok {
		assert.Fail(t, fmt.Sprintf("Expected ok handle, but it's failed: %s", handle.String()))
	}
	if _, ok := c.checkedIndexes[index]; ok {
		assert.Fail(t, fmt.Sprintf("Handle already checked: %s", handle.String()))
	}
	c.checkedIndexes[index] = struct{}{}
	return handle
}

func (c *skrStatusHandleCheckerImpl) Ok(val bool) SkrStatusHandleChecker {
	c.nok = !val
	return c
}

func (c *skrStatusHandleCheckerImpl) Name(v string) SkrStatusHandleChecker {
	c.objName = v
	return c
}

func (c *skrStatusHandleCheckerImpl) Obj(v string) SkrStatusHandleChecker {
	c.obj = v
	c.crd = ""
	c.busola = ""
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
