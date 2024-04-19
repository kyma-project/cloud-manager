package config

type PathBuilder struct {
	cfg  Config
	path FieldPath
}

func (b *PathBuilder) DefaultScalar(scalar interface{}) *PathBuilder {
	b.cfg.DefaultScalar(b.path, scalar)
	return b
}

func (b *PathBuilder) DefaultObj(obj interface{}) *PathBuilder {
	b.cfg.DefaultObj(b.path, obj)
	return b
}

func (b *PathBuilder) DefaultJson(js string) *PathBuilder {
	b.cfg.DefaultJson(b.path, js)
	return b
}

func (b *PathBuilder) SourceFile(file string) *PathBuilder {
	b.cfg.SourceFile(b.path, file)
	return b
}

func (b *PathBuilder) SourceEnv(envVarPrefix string) *PathBuilder {
	b.cfg.SourceEnv(b.path, envVarPrefix)
	return b
}

func (b *PathBuilder) Bind(obj interface{}) *PathBuilder {
	b.cfg.Bind(b.path, obj)
	return b
}
