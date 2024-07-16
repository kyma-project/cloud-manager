package config

import (
	"github.com/fsnotify/fsnotify"
)

type Config interface {
	BaseDir(baseDir string)
	GetEnv(key string) string
	Path(path string, actions ...PathAction) Config
	DefaultScalar(path string, scalar interface{})
	DefaultObj(path string, obj interface{})
	DefaultJson(path string, js string)
	SourceFile(fieldPath string, file string)
	SourceEnv(fieldPath string, envVarPrefix string)
	Sensitive(fieldPath string)
	Read()
	Watch(stopCh <-chan struct{}, onConfigChange func(event fsnotify.Event)) error
	Bind(fieldPath string, obj any)
	Json() string
	PrintJson() string
	GetAsString(path string) string
}

type Source interface {
	Read(in string) string
}

type Binding interface {
	Copy(in string)
}
