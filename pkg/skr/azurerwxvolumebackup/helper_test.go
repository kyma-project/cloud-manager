package azurerwxvolumebackup

import "context"

type ContextKeyTypes interface {
	string | int
}

type ContextValueTypes interface {
	string | int | bool
}

func addValuesToContext[K ContextKeyTypes, V ContextValueTypes](ctx context.Context, kvp map[K]V) context.Context {

	for k, v := range kvp {
		ctx = context.WithValue(ctx, k, v)
	}
	return ctx

}
