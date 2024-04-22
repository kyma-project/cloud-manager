package config

type PathBuilder struct {
	cfg  Config
	path string
}

type PathAction func(cfg Config, path string)

func (b *PathBuilder) Apply(actions ...PathAction) {
	for _, a := range actions {
		a(b.cfg, b.path)
	}
}

func Path(subPath string, actions ...PathAction) PathAction {
	return func(cfg Config, path string) {
		thePath := ConcatFieldPath(path, subPath)
		for _, a := range actions {
			a(cfg, thePath)
		}
	}
}

//func x() *PathBuilder {
//	return &PathBuilder{
//		cfg:  b.cfg,
//		path: ConcatFieldPath(b.path, subPath),
//	}
//}

func DefaultScalar(scalar interface{}) PathAction {
	return func(cfg Config, path string) {
		cfg.DefaultScalar(path, scalar)
	}
}

func DefaultObj(obj interface{}) PathAction {
	return func(cfg Config, path string) {
		cfg.DefaultObj(path, obj)
	}
}

func DefaultJson(js string) PathAction {
	return func(cfg Config, path string) {
		cfg.DefaultJson(path, js)
	}
}

func SourceFile(file string) PathAction {
	return func(cfg Config, path string) {
		cfg.SourceFile(path, file)
	}
}

func SourceEnv(envVarPrefix string) PathAction {
	return func(cfg Config, path string) {
		cfg.SourceEnv(path, envVarPrefix)
	}
}

func Bind(obj interface{}) PathAction {
	return func(cfg Config, path string) {
		cfg.Bind(path, obj)
	}
}

func Sensitive() PathAction {
	return func(cfg Config, path string) {
		cfg.Sensitive(path)
	}
}
