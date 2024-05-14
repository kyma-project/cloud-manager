package config

import (
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"testing"
)

type jsonAssertion struct {
	path string
	tp   gjson.Type
	str  string
}

func TestRawConfigAtPathFromOneYamlFile(t *testing.T) {
	env := abstractions.NewMockedEnvironment(nil)
	cfg := NewConfig(env)
	cfg.SourceFile("root", "testdata/first.yaml")
	cfg.Read()

	expected := []jsonAssertion{
		{path: "root", tp: gjson.JSON},
		{path: "root.string", tp: gjson.String, str: "abc"},
		{path: "root.number", tp: gjson.Number, str: "23"},
		{path: "root.arrStr", tp: gjson.JSON},
		{path: "root.arrStr.0", tp: gjson.String, str: "a"},
		{path: "root.arrNum", tp: gjson.JSON},
		{path: "root.arrNum.0", tp: gjson.Number, str: "1"},
		{path: "root.map.nested.x", tp: gjson.String, str: "xxx"},
	}

	for _, ex := range expected {
		t.Run(ex.path, func(t *testing.T) {
			res := gjson.Get(cfg.Json(), ex.path)
			assert.Equal(t, ex.tp, res.Type, "expected %s to be %s but it is %s", ex.path, ex.tp, res.Type)
			if ex.str != "" {
				assert.Equal(t, ex.str, res.String(), "expected %s to equal %s but it is %s", ex.path, ex.str, res.String())
			}
		})
	}
}

func TestRawConfigAtPathFromOneJsonFile(t *testing.T) {
	env := abstractions.NewMockedEnvironment(nil)
	cfg := NewConfig(env)
	cfg.SourceFile("root", "testdata/first.json")
	cfg.Read()

	expected := []jsonAssertion{
		{path: "root", tp: gjson.JSON},
		{path: "root.string", tp: gjson.String, str: "abc"},
		{path: "root.number", tp: gjson.Number, str: "23"},
		{path: "root.arrStr", tp: gjson.JSON},
		{path: "root.arrStr.0", tp: gjson.String, str: "a"},
		{path: "root.arrNum", tp: gjson.JSON},
		{path: "root.arrNum.0", tp: gjson.Number, str: "1"},
		{path: "root.map.nested.x", tp: gjson.String, str: "xxx"},
	}

	for _, ex := range expected {
		t.Run(ex.path, func(t *testing.T) {
			res := gjson.Get(cfg.Json(), ex.path)
			assert.Equal(t, ex.tp, res.Type, "expected %s to be %s but it is %s", ex.path, ex.tp, res.Type)
			if ex.str != "" {
				assert.Equal(t, ex.str, res.String(), "expected %s to equal %s but it is %s", ex.path, ex.str, res.String())
			}
		})
	}
}

func TestRawConfigAtPathFromEnv(t *testing.T) {
	env := abstractions.NewMockedEnvironment(map[string]string{
		"FOO":               "bar",
		"TEST_A":            "a",
		"TEST_X__Y":         "x.y",
		"TEST_FOO_BAR__BAZ": "fooBar.baz",
	})
	cfg := NewConfig(env)
	cfg.SourceEnv("root", "TEST_")
	cfg.Read()

	expected := []jsonAssertion{
		{path: "root", tp: gjson.JSON},
		{path: "root.a", tp: gjson.String, str: "a"},
		{path: "root.x.y", tp: gjson.String, str: "x.y"},
		{path: "root.fooBar.baz", tp: gjson.String, str: "fooBar.baz"},
	}

	for _, ex := range expected {
		t.Run(ex.path, func(t *testing.T) {
			res := gjson.Get(cfg.Json(), ex.path)
			assert.Equal(t, ex.tp, res.Type, "expected %s to be %s but it is %s", ex.path, ex.tp, res.Type)
			if ex.str != "" {
				assert.Equal(t, ex.str, res.String(), "expected %s to equal %s but it is %s", ex.path, ex.str, res.String())
			}
		})
	}
}

func TestRawConfigInRootFromOneYamlFile(t *testing.T) {
	env := abstractions.NewMockedEnvironment(nil)
	cfg := NewConfig(env)
	cfg.SourceFile("", "testdata/first.yaml")
	cfg.Read()

	expected := []jsonAssertion{
		{path: "string", tp: gjson.String, str: "abc"},
		{path: "number", tp: gjson.Number, str: "23"},
		{path: "arrStr", tp: gjson.JSON},
		{path: "arrStr.0", tp: gjson.String, str: "a"},
		{path: "arrNum", tp: gjson.JSON},
		{path: "arrNum.0", tp: gjson.Number, str: "1"},
		{path: "map.nested.x", tp: gjson.String, str: "xxx"},
	}

	for _, ex := range expected {
		t.Run(ex.path, func(t *testing.T) {
			res := gjson.Get(cfg.Json(), ex.path)
			assert.Equal(t, ex.tp, res.Type, "expected %s to be %s but it is %s", ex.path, ex.tp, res.Type)
			if ex.str != "" {
				assert.Equal(t, ex.str, res.String(), "expected %s to equal %s but it is %s", ex.path, ex.str, res.String())
			}
		})
	}
}

func TestRawConfigInRootFromTwoMergedYamlFiles(t *testing.T) {
	env := abstractions.NewMockedEnvironment(nil)
	cfg := NewConfig(env)
	cfg.SourceFile("", "testdata/resourceQuotaDefault.yaml")
	cfg.SourceFile("", "testdata/resourceQuotaOverride.yaml")
	cfg.Read()

	expected := []jsonAssertion{
		{path: "resourceQuota.skr.defaults.ipranges", tp: gjson.Number, str: "2"},
		{path: "resourceQuota.skr.defaults.awsnfsvolumes", tp: gjson.Number, str: "3"},
		{path: "resourceQuota.skr.overrides.skr123.ipranges", tp: gjson.Number, str: "2"},
		{path: "resourceQuota.skr.overrides.skr789.ipranges", tp: gjson.Number, str: "4"},
	}

	for _, ex := range expected {
		t.Run(ex.path, func(t *testing.T) {
			res := gjson.Get(cfg.Json(), ex.path)
			assert.Equal(t, ex.tp, res.Type, "expected %s to be %s but it is %s", ex.path, ex.tp, res.Type)
			if ex.str != "" {
				assert.Equal(t, ex.str, res.String(), "expected %s to equal %s but it is %s", ex.path, ex.str, res.String())
			}
		})
	}
}

func TestRawConfigInRootFromTwoMergedJsonFiles(t *testing.T) {
	env := abstractions.NewMockedEnvironment(nil)
	cfg := NewConfig(env)
	cfg.SourceFile("", "testdata/resourceQuotaDefault.json")
	cfg.SourceFile("", "testdata/resourceQuotaOverride.json")
	cfg.Read()

	expected := []jsonAssertion{
		{path: "resourceQuota.skr.defaults.ipranges", tp: gjson.Number, str: "2"},
		{path: "resourceQuota.skr.defaults.awsnfsvolumes", tp: gjson.Number, str: "3"},
		{path: "resourceQuota.skr.overrides.skr123.ipranges", tp: gjson.Number, str: "2"},
		{path: "resourceQuota.skr.overrides.skr789.ipranges", tp: gjson.Number, str: "4"},
	}

	for _, ex := range expected {
		t.Run(ex.path, func(t *testing.T) {
			res := gjson.Get(cfg.Json(), ex.path)
			assert.Equal(t, ex.tp, res.Type, "expected %s to be %s but it is %s", ex.path, ex.tp, res.Type)
			if ex.str != "" {
				assert.Equal(t, ex.str, res.String(), "expected %s to equal %s but it is %s", ex.path, ex.str, res.String())
			}
		})
	}
}

func TestRawConfigInRootFromScalarFile(t *testing.T) {
	cfg := NewConfig(abstractions.NewMockedEnvironment(nil))
	cfg.SourceFile("aaa.bbb", "testdata/scalar")
	cfg.Read()

	res := gjson.Get(cfg.Json(), "aaa.bbb")
	assert.Equal(t, gjson.String, res.Type)
	assert.Equal(t, "some value", res.String())
}
