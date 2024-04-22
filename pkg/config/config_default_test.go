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
	cfg.DefaultScalar(NewFiledPath("scalar", "path"), "scalarValue")
	cfg.DefaultJson(NewFiledPath("json", "path"), "{\"arrNum\":[1,2],\"arrStr\":[\"a\",\"b\"],\"map\":{\"a\":\"aaa\",\"b\":\"bbb\",\"nested\":{\"x\":\"xxx\",\"y\":\"yyy\"}},\"number\":23,\"string\":\"abc\"}")
	cfg.DefaultObj(NewFiledPath("obj", "path"), &SomeStruct{Foo: "foo", Bar: 32})
	cfg.Read()

	assert.Equal(t, "scalarValue", cfg.GetAsString(NewFiledPath("scalar", "path")))
	assert.Equal(t, "1", cfg.GetAsString(NewFiledPath("json", "path", "arrNum", "0")))
	assert.Equal(t, "yyy", cfg.GetAsString(NewFiledPath("json", "path", "map", "nested", "y")))
	assert.Equal(t, "foo", cfg.GetAsString(NewFiledPath("obj", "path", "fooooo")))
	assert.Equal(t, "32", cfg.GetAsString(NewFiledPath("obj", "path", "Bar")))
}

func TestConfigDefaultMergedWithLoaded(t *testing.T) {
	cfg := NewConfig(abstractions.NewMockedEnvironment(map[string]string{
		"TEST_FOO_BAR__BAZ": "fooBar.baz",
	}))

	cfg.DefaultScalar(NewFiledPath("root", "map", "nested", "z"), "zzz")
	cfg.DefaultScalar(NewFiledPath("root", "map", "nested", "x"), "xxxxxxxxx")
	cfg.DefaultScalar(NewFiledPath("root", "default"), "dummy")
	cfg.DefaultScalar(NewFiledPath("root", "fooBar", "some"), "some")
	cfg.DefaultScalar(NewFiledPath("root", "arrStr", "0"), "c")

	cfg.SourceFile(NewFiledPath("root"), "testdata/first.yaml")
	cfg.SourceEnv(NewFiledPath("root"), "TEST_")

	cfg.Read()

	// Default values that are NOT specified in the sources still are available
	assert.Equal(t, "zzz", cfg.GetAsString(NewFiledPath("root", "map", "nested", "z")))
	assert.Equal(t, "dummy", cfg.GetAsString(NewFiledPath("root", "default")))
	assert.Equal(t, "some", cfg.GetAsString(NewFiledPath("root", "fooBar", "some")))

	// Default values that are specified in the sources are overwritten
	// whole array is replaced, and default is overwritten
	assert.NotEqual(t, "c", cfg.GetAsString(NewFiledPath("root", "arrStr", "0")))
	assert.NotEqual(t, "xxxxxxxxx", cfg.GetAsString(NewFiledPath("root", "map", "nested", "x")))

	// Values from sources
	assert.Equal(t, "xxx", cfg.GetAsString(NewFiledPath("root", "map", "nested", "x")))
	assert.Equal(t, "fooBar.baz", cfg.GetAsString(NewFiledPath("root", "fooBar", "baz")))
}
