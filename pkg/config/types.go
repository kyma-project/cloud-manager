package config

import "github.com/fsnotify/fsnotify"

type Config interface {
	SourceFile(fieldPath FieldPath, file string)
	SourceEnv(fieldPath FieldPath, envVarPrefix string)
	Read()
	Watch(stopCh chan struct{}, onConfigChange func(event fsnotify.Event)) error
	Bind(fieldPath FieldPath, obj any)
	Json() string
}

type Source interface {
	Read(in string) string
}

type Binding interface {
	Copy(raw map[string]interface{})
}
