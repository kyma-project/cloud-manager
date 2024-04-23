package config

import (
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfigDefaults(t *testing.T) {
	type SomeStruct struct {
		Foo string `json:"fooooo"`
		Bar int
	}
	cfg := NewConfig(abstractions.NewMockedEnvironment(nil))
	cfg.DefaultScalar("scalar.path", "scalarValue")
	cfg.DefaultJson("json.path", "{\"arrNum\":[1,2],\"arrStr\":[\"a\",\"b\"],\"map\":{\"a\":\"aaa\",\"b\":\"bbb\",\"nested\":{\"x\":\"xxx\",\"y\":\"yyy\"}},\"number\":23,\"string\":\"abc\"}")
	cfg.DefaultObj("obj.path", &SomeStruct{Foo: "foo", Bar: 32})
	cfg.Read()

	assert.Equal(t, "scalarValue", cfg.GetAsString("scalar.path"))
	assert.Equal(t, "1", cfg.GetAsString("json.path.arrNum.0"))
	assert.Equal(t, "yyy", cfg.GetAsString("json.path.map.nested.y"))
	assert.Equal(t, "foo", cfg.GetAsString("obj.path.fooooo"))
	assert.Equal(t, "32", cfg.GetAsString("obj.path.Bar"))
}

func TestConfigDefaultMergedWithLoaded(t *testing.T) {
	cfg := NewConfig(abstractions.NewMockedEnvironment(map[string]string{
		"TEST_FOO_BAR__BAZ": "fooBar.baz",
	}))

	cfg.DefaultScalar("root.map.nested.z", "zzz")
	cfg.DefaultScalar("root.map.nested.x", "xxxxxxxxx")
	cfg.DefaultScalar("root.default", "dummy")
	cfg.DefaultScalar("root.fooBar.some", "some")
	cfg.DefaultScalar("root.arrStr.0", "c")

	cfg.SourceFile("root", "testdata/first.yaml")
	cfg.SourceEnv("root", "TEST_")

	cfg.Read()

	// Default values that are NOT specified in the sources still are available
	assert.Equal(t, "zzz", cfg.GetAsString("root.map.nested.z"))
	assert.Equal(t, "dummy", cfg.GetAsString("root.default"))
	assert.Equal(t, "some", cfg.GetAsString("root.fooBar.some"))

	// Default values that are specified in the sources are overwritten
	// whole array is replaced, and default is overwritten
	assert.NotEqual(t, "c", cfg.GetAsString("root.arrStr.0"))
	assert.NotEqual(t, "xxxxxxxxx", cfg.GetAsString("root.arrStr.0"))

	// Values from sources
	assert.Equal(t, "xxx", cfg.GetAsString("root.map.nested.x"))
	assert.Equal(t, "fooBar.baz", cfg.GetAsString("root.fooBar.baz"))
}
