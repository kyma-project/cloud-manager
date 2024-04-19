# Config

Main features:
* Internal CloudManager usage
* Files and environment variables as sources
* Supported source file encoding formats: YAML, JSON, scalar values
* Multiple sources merged into single configuration at predefined field paths
* Filesystem watch and automatic reload on file changed or symlink replaced (ie: k8s ConfigMap change)
* Bound Golang structs at a specified field path with automatic reload on the config change
* Mapping from raw config to a bound struct with `github.com/mitchellh/mapstructure`


## Usage

```golang
package main

import (
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/config"
	"github.com/tidwall/gjson"
)

type someStruct struct {
	Field string `mapstructure:"fieldName"`
}

func main() {
	cfg := config.NewConfig(abstractions.NewOSEnvironment())
	cfg.SourceFile(config.NewFiledPath("some.path"), "/path/to/some/file.yaml")
	cfg.SourceFile(config.NewFiledPath("other.path"), "/path/to/other/file.json")
	cfg.SourceFile(config.NewFiledPath("third.path"), "/path/to/scalar/value")
	obj := new(someStruct)
	cfg.Bind(config.NewFiledPath("raw.path.to.bound"), obj)
	cfg.Read()
	if err := cfg.Watch(nil, nil); err != nil {
		panic(err)
	}

	// raw string access
	a := gjson.Get(cfg.Json(), "some.path.filed").String()
	
	// access trough bound struct
	b := obj.Field
}
```
