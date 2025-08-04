package e2e

import (
	"context"
)

type World struct {
	ClusterProvider ClusterProvider
}

type worldKey struct{}

func setWorld(ctx context.Context, fc *World) context.Context {
	return context.WithValue(ctx, worldKey{}, fc)
}

func getWorld(ctx context.Context) *World {
	g, _ := ctx.Value(worldKey{}).(*World)

	return g
}

const (
	defaultKcpNamespace = "kcp-system"
)

func NewWorld() *World {
	return &World{
		ClusterProvider: &defaultClusterProvider{kcpNamespace: defaultKcpNamespace},
	}
}

func (w *World) KcpNamespace() string {
	return defaultKcpNamespace
}

func (w *World) GardenNamespace() string{
	return "garden"
}

func (w *World) SkrNamespace() string {
	return "skr"
}
