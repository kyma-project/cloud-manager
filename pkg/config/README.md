# Config

Main features:
* Internal CloudManager usage
* Files and environment variables as sources
* Supported source file encoding formats: yaml, json, scalar values
* Multiple sources merged into single configuration at predefined field paths
* Filesystem watch and automatic reload on file changed or symlink replaced (ie: k8s configmap change)
* Bound golang structs at specified field path with automatic reload on config change
* Mapping from raw config to bound struct with `github.com/mitchellh/mapstructure`


# Usage

```golang
package main

import (
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/config"
)

type someData struct {
	Field string `mapstructure:"fieldName"`
}

func main()  {
	cfg := config.NewConfig(abstractions.NewOSEnvironment())
	cfg.SourceFile(config.NewFiledPath("some.path"), "/path/to/some/file.yaml")
	cfg.SourceFile(config.NewFiledPath("other.path"), "/path/to/other/file.json")
	cfg.SourceFile(config.NewFiledPath("third.path"), "/path/to/scalar/value")
	obj := &someData{}
	cfg.Bind(config.NewFiledPath("raw.path.to.bound.struct"), obj)
    cfg.Read()
	if err := cfg.Watch(nil, nil); err != nil {
		panic(err)
    }

}

```