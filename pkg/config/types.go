package config

import (
	"github.com/fsnotify/fsnotify"
)

type Config interface {
	Path(path FieldPath) *PathBuilder
	DefaultScalar(path FieldPath, scalar interface{})
	DefaultObj(path FieldPath, obj interface{})
	DefaultJson(path FieldPath, js string)
	SourceFile(fieldPath FieldPath, file string)
	SourceEnv(fieldPath FieldPath, envVarPrefix string)
	Read()
	Watch(stopCh chan struct{}, onConfigChange func(event fsnotify.Event)) error
	Bind(fieldPath FieldPath, obj any)
	Json() string
	GetAsString(path FieldPath) string
}

type Source interface {
	Read(in string) string
}

type Binding interface {
	Copy(raw map[string]interface{})
}
